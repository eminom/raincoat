* Processing a raw data directly, in pb-mode

  dmaster topspti.raw.data

* Processing a raw dpf ring buffer dump file directly

  dmaster -rawdpf -t20 0_cluster.bin

* Dump meta infomration from a raw data in pb-mode

  dmaster -dumpmeta topspti.raw.data

* Dump DPF events fro a raw data in pb-mode

  dmaster -dump topspti.raw.data
