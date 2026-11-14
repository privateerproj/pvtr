# Plugin Supply-Chain Security

This page explains how pvtr checks that a plugin binary came from a known
publisher and was not changed in transit. It covers how plugins are signed,
what pvtr verifies when it installs a plugin, where trust on first use leaves
gaps, and how to install from an internal mirror with no internet access.

## How plugins are signed

`pvtr publish` pushes a plugin release to grc.store, then signs it
**keylessly** with [Sigstore](https://www.sigstore.dev/) and attaches the
signature to the pushed index:

- The signer gets a short-lived Fulcio certificate tied to an OIDC identity,
  such as the GitHub Actions workflow that built the release. No long-lived
  private key has to be stored, rotated, or published.
- The signature covers the plugin's OCI image-index digest. That digest in turn
  pins the per-platform manifests, config, and binary layers.
- The signature is recorded in the public Rekor transparency log. The resulting
  Sigstore bundle is attached to the index as an OCI 1.1 referrer.

`pvtr publish --registry …` and `pvtr publish --no-sync` skip signing. Their
output is unsigned and fails verification at install.

Because signing is keyless, there is no publisher key to publish. A plugin is
trusted through its **signer identity**, in the form
`keyless:<oidc-issuer>#<subject>`. Every signature can be audited in the public
transparency log.

The subject is the certificate's subject (SAN) with its per-release
`@refs/...` suffix removed, so the identity names the workflow file, not one
tag, and stays the same across releases. For a GitHub Actions release, the
certificate subject

```text
https://github.com/acme/pvtr-plugin/.github/workflows/release.yml@refs/tags/v1.2.0
```

is recorded as

```text
keyless:https://token.actions.githubusercontent.com#https://github.com/acme/pvtr-plugin/.github/workflows/release.yml
```

Compare against this form, not the full certificate subject.

## What pvtr verifies at install

`pvtr install <namespace>/<plugin_id>[@<version>]` runs these checks before it
writes the binary. If a check fails, the install stops. It never falls back to
an unverified copy.

1. **Signature.** The Sigstore bundle must chain to the Sigstore public-good
   trusted root, which is built into pvtr. That means a valid Fulcio
   certificate, an embedded SCT, and a Rekor inclusion proof. This check runs
   offline.
2. **Signer identity.** See [Trust on first use](#trust-on-first-use) for
   exactly which identity is enforced.
3. **Digest chain.** pvtr checks every step from the signed index to the
   platform manifest, the config, and the binary bytes against its sha256
   digest.
4. **Hub/registry agreement.** When the hub has recorded an index digest for
   the release, the index pulled from the registry must match it. If the hub
   has no digest recorded, pvtr prints a warning, skips this check, and
   continues.

After writing the binary, `pvtr install` also installs every catalog the plugin
declares, through the same hub and registry. Each catalog's signature is
verified against the signer identity the hub declares for that catalog, or
accepted from any valid keyless signer when the hub declares none. Catalogs
have no local pin.

The verified index digest and signer identity are recorded in
`<binaries-path>/plugins.json`. `pvtr list` does not show them, so check that
file directly.

### Trust on first use

The signer identity is pinned per plugin, and the pin is chosen like this:

- **The plugin is already installed:** the identity pinned in your local
  `plugins.json` is enforced. Updates must be signed by the same identity, and
  this local pin wins over whatever the hub declares.
- **First install, and the hub declares an identity:** that identity is
  enforced and becomes the pin. The hub is choosing the signer you trust.
- **First install, and the hub declares no identity:** any valid keyless
  signature is accepted, and its identity becomes the pin.

A fresh CI runner, or any new `--binaries-path`, is always a first install. So
in those environments, the trust decision rests with the hub you point pvtr at.
To keep a known identity across fresh environments, carry the `plugins.json`
from a trusted install forward with the binaries, or compare the signer
identity in `plugins.json` with the one you expect after installing.

## Air-gapped and internal-registry distribution

pvtr connects to two endpoints:

- **The hub,** set by `hub-url` (default `https://hub.grc.store`).
- **The OCI registry the hub advertises** as `registry_url` in its
  `/.well-known/grc-store-configuration` document (`https://oci.grc.store` for
  the public hub). It can be on any host.

Both must be reachable, so a firewall allowlist needs both hosts. To install
without reaching the public internet:

1. **Run an internal grc.store-compatible hub and OCI registry.** The hub must
   serve the well-known document (advertising your internal registry), the
   plugin-detail API (`/v1/plugins/<namespace>/<plugin_id>`), and the
   catalog-detail API (`/v1/catalogs/<namespace>/<catalog_id>`). Without the
   catalog-detail API, pvtr treats every catalog as unpublished: the install
   only warns and reports success with no catalogs cached.

   pvtr reads from both anonymously and has no way to pass credentials, so the
   hub and the registry must allow unauthenticated reads of the mirrored
   repositories.
2. **Mirror the signed artifacts, signatures included.** Copy each plugin
   index, its child manifests and blobs, **and its referrers** (the Sigstore
   bundles). Plugin repositories sit under a reserved `plugins` path segment.
   For example, with [ORAS](https://oras.land/):

   ```bash
   oras cp -r oci.grc.store/<namespace>/plugins/<plugin_id>:<version> \
     registry.internal.example/<namespace>/plugins/<plugin_id>:<version>
   ```

3. **Mirror the catalogs each plugin declares.** Catalog repositories have no
   reserved segment:

   ```bash
   oras cp -r oci.grc.store/<namespace>/<catalog_id>:<version> \
     registry.internal.example/<namespace>/<catalog_id>:<version>
   ```

   Which side is missing decides what happens:

   - **The hub has no record of the catalog, or of that version:** a warning
     only. The install succeeds and the plugin must carry its own copy.
   - **The hub lists the catalog, but the registry is missing the repository
     or tag, or the catalog fails verification:** `pvtr install` returns an
     error, but only after it has already written the plugin binary and its
     `plugins.json` entry.
4. **Point pvtr at the internal hub:**

   ```bash
   export PVTR_HUB_URL=https://grc-hub.internal.example
   pvtr install <namespace>/<plugin_id>@<version>
   ```

Copying by digest keeps the index digest unchanged, so the original signature
still verifies. Do **not** rebuild, re-tag by content, or re-sign mirrored
artifacts.

Signature verification needs no network access. The Sigstore trusted root is
built into pvtr, and each bundle carries its own transparency-log proof. Which
identity is enforced follows the [trust on first use](#trust-on-first-use)
rules above. On a first install, that means the identity your **internal**
hub declares. Make sure the internal hub declares each plugin's original
publisher identity, or carry `plugins.json` forward from a trusted install.

When the public Sigstore trusted root rotates, upgrade pvtr to pick up the new
root.
