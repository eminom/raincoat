package main

import "git.enflame.cn/hai.bai/dmaster/codec"

func main() {
	check := codec.DoradoMidCheckout{}
	for _, target := range []string{
		codec.ENGINE_PCIE,
		codec.ENGINE_HCVG,
		codec.ENGINE_VDEC,
		codec.ENGINE_TS,
	} {
		check.CheckoutFor(target)
	}
}
