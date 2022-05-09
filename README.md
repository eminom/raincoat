* Processing a raw data directly, in pb-mode

```bash
  dmaster topspti.raw.data
```

* Processing a raw dpf ring buffer dump file directly

```bash
  dmaster -rawdpf -t20 0_cluster.bin
```

* Dump meta infomration from a raw data in pb-mode

```bash
  dmaster -dumpmeta topspti.raw.data
```

* Dump DPF events fro a raw data in pb-mode

```bash
  dmaster -dump topspti.raw.data
```

* Check executable's profile section

```bash
  dmaster -exec -checkf sip.bin
```
