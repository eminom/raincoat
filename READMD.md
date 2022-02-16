* Processing a raw data directly, in pb-mode

  dmaster -pb topspti.raw.data

* Processing a raw dpf ring buffer dump file directly

  dmaster -dump -raw -meta $(pwd) -job 0 -arch pavo 0_cluster.bin

* Dump meta infomration from a raw data in pb-mode

  dmaster -pb -dumpmeta topspti.raw.data

* Dump DPF events fro a raw data in pb-mode

  dmaster -pb -dump topspti.raw.data
