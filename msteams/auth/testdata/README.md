# Synthetic certificate fixtures

Both PFX files contain the same generated 2048-bit RSA key and a self-signed certificate
with subject `Synthetic SDK test only`, serial 81002 and a 2026-01-01 through 2036-01-01
validity window. The password is the public test value `fixture-only`. These credentials
have never been registered in a tenant and must never be used for authentication.
Tests use a fixed date inside the window, so their result is independent of calendar time.

`client-credentials-sha1.pfx` exercises the PKCS#12 subset supported by the pinned Azure
Identity parser (SHA1 MAC and 3DES PBE). `client-credentials-sha256.pfx` exercises rejection
of a SHA256 MAC/AES bundle. This is an interoperability test, not a recommendation to
create legacy-encrypted production bundles; unencrypted PEM is the portable alternative.

Generated locally with Go crypto/rand, crypto/rsa and crypto/x509. Packaged with OpenSSL
`pkcs12 -export`, using explicit `-keypbe`/`-certpbe` values `PBE-SHA1-3DES` or `AES-256-CBC`
and `-macalg sha1` or `sha256`. OpenSSL is not needed when running the tests.
