package keeper

////  sP
//func calcSpotPrice(
//	tokenBalanceIn,
//	tokenWeightIn,
//	tokenBalanceOut,
//	tokenWeightOut,
//	swapFee sdk.Dec,
//) sdk.Dec {
//	number := tokenBalanceIn.Quo(tokenWeightIn)
//	denom := tokenBalanceOut.Quo(tokenWeightOut)
//	ratio := number.Quo(denom)
//	scale := sdk.OneDec().Quo(sdk.OneDec().Sub(swapFee))
//
//	return ratio.Mul(scale)
//}
//
//// aO
//func calcOutGivenIn(
//	tokenBalanceIn,
//	tokenWeightIn,
//	tokenBalanceOut,
//	tokenWeightOut,
//	tokenAmountIn,
//	swapFee sdk.Dec,
//) sdk.Dec {
//	weightRatio := tokenWeightIn.Quo(tokenWeightOut)
//	adjustedIn := sdk.OneDec().Sub(swapFee)
//	adjustedIn := tokenAmountIn.Mul(adjustedIn)
//	y := tokenBalanceIn.Quo(tokenBalanceIn.Add(adjustedIn))
//	foo := y.Power(weightRatio)
//	bar := sdk.OneDec().Sub(foo)
//	return tokenBalanceOut.Mul(bar)
//}
//
//// aI
//func calcInGivenOut(
//	tokenBalanceIn,
//	tokenWeightIn,
//	tokenBalanceOut,
//	tokenWeightOut,
//	tokenAmountOut,
//	swapFee sdk.Dec,
//) sdk.Dec {
//	weightRatio := tokenWeightOut.Quo(tokenWeightIn)
//	diff := tokenBalanceOut.Sub(tokenAmountOut)
//	y := tokenBalanceOut.Quo(diff)
//	foo := y.Power(weightRatio)
//	foo := foo.Sub(sdk.OneDec())
//	tokenAmountIn := sdk.OneDec().Sub(swapFee)
//	return (tokenBalanceIn.Mul(foo)).Quo(tokenAmountIn)
//
//}
//
//// pAo
//func calcPoolOutGivenSingleIn(
//	tokenBalanceIn,
//	tokenWeightIn,
//	poolSupply,
//	totalWeight,
//	tokenAmountIn,
//	swapFee sdk.Dec,
//) sdk.Dec {
//	normalizedWeight := tokenWeightIn.Quo(totalWeight)
//	zaz := (sdk.OneDec().Sub(normalizedWeight)).Mul(swapFee)
//	tokenAmountInAfterFee := tokenAmountIn.Mul(sdk.OneDec().Sub(zaz))
//
//	newTokenBalanceIn := tokenBalanceIn.Add(tokenAmountInAfterFee)
//	tokenInRatio := newTokenBalanceIn.Quo(tokenBalanceIn)
//
//	// uint newPoolSupply = (ratioTi ^ weightTi) * poolSupply;
//	poolRatio := tokenInRatio.Power(normalizedWeight)
//	newPoolSupply := poolRatio.Mul(poolSupply)
//	return newPoolSupply.Sub(poolSupply)
//}
//
////tAi
//func calcSingleInGivenPoolOut(
//	tokenBalanceIn,
//	tokenWeightIn,
//	poolSupply,
//	totalWeight,
//	poolAmountOut,
//	swapFee sdk.Dec,
//) sdk.Dec {
//	normalizedWeight := tokenWeightIn.Quo(totalWeight)
//	newPoolSupply := poolSupply.Add(poolAmountOut)
//	poolRatio := newPoolSupply.Div(poolSupply)
//
//	//uint newBalTi = poolRatio^(1/weightTi) * balTi;
//	boo := sdk.OneDec().Quo(normalizedWeight)
//	tokenInRatio := poolRatio.Power(boo)
//	newTokenBalanceIn := tokenInRatio.Mul(tokenBalanceIn)
//	tokenAmountInAfterFee := newTokenBalanceIn.Sub(tokenBalanceIn)
//	// Do reverse order of fees charged in joinswap_ExternAmountIn, this way
//	//     ``` pAo == joinswap_ExternAmountIn(Ti, joinswap_PoolAmountOut(pAo, Ti)) ```
//	//uint tAi = tAiAfterFee / (1 - (1-weightTi) * swapFee) ;
//	zar := (sdk.OneDec().Sub(normalizedWeight)).Mul(swapFee)
//	return tokenAmountInAfterFee.Quo(sdk.OneDec().Sub(zar))
//}
//
//// tAo
//func calcSingleOutGivenPoolIn(
//	tokenBalanceOut,
//	tokenWeightOut,
//	poolSupply,
//	totalWeight,
//	poolAmountIn,
//	swapFee sdk.Dec,
//) sdk.Dec {
//	normalizedWeight := tokenWeightOut.Quo(totalWeight)
//	// charge exit fee on the pool token side
//	// pAiAfterExitFee = pAi*(1-exitFee)
//	poolAmountInAfterExitFee := poolAmountIn.Mul(sdk.OneDec())
//	newPoolSupply := poolSupply.Sub(poolAmountInAfterExitFee)
//	poolRatio := newPoolSupply.Quo(poolSupply)
//
//	// newBalTo = poolRatio^(1/weightTo) * balTo;
//
//	tokenOutRatio := poolRatio.Power(sdk.OneDec().Quo(normalizedWeight))
//	newTokenBalanceOut := tokenOutRatio.Mul(tokenBalanceOut)
//
//	tokenAmountOutBeforeSwapFee := tokenBalanceOut.Sub(newTokenBalanceOut)
//
//	// charge swap fee on the output token side
//	//uint tAo = tAoBeforeSwapFee * (1 - (1-weightTo) * swapFee)
//	zaz := (sdk.OneDec().Sub(normalizedWeight)).Mul(swapFee)
//	tokenAmountOut := tokenAmountOutBeforeSwapFee.Mul(sdk.OneDec().Sub(zaz))
//	return tokenAmountOut
//}
//
//// pAi
//func calcPoolInGivenSingleOut(
//	tokenBalanceOut,
//	tokenWeightOut,
//	poolSupply,
//	totalWeight,
//	tokenAmountOut,
//	swapFee sdk.Dec,
//) sdk.Dec {
//	// charge swap fee on the output token side
//	normalizedWeight := tokenWeightOut.Quo(totalWeight)
//	//uint tAoBeforeSwapFee = tAo / (1 - (1-weightTo) * swapFee) ;
//	zoo := sdk.OneDec().Sub(normalizedWeight)
//	zar := zoo.Mul(swapFee)
//	tokenAmountOutBeforeSwapFee := tokenAmountOut.Quo(sdk.OneDec().Sub(zar))
//
//	newTokenBalanceOut := tokenBalanceOut.Sub(tokenAmountOutBeforeSwapFee)
//	tokenOutRatio := newTokenBalanceOut.Quo(tokenBalanceOut)
//
//	//uint newPoolSupply = (ratioTo ^ weightTo) * poolSupply;
//	poolRatio := tokenOutRatio.Power(normalizedWeight)
//	newPoolSupply := poolRatio.Mul(poolSupply)
//	poolAmountInAfterExitFee := poolSupply.Sub(newPoolSupply)
//
//	// charge exit fee on the pool token side
//	// pAi = pAiAfterExitFee/(1-exitFee)
//	return poolAmountInAfterExitFee.Quo(sdk.OneDec())
//}
//
//
//
///*
//function bsubSign(uint a, uint b)
//internal pure
//returns (uint, bool)
//{
//if (a >= b) {
//return (a - b, false);
//} else {
//return (b - a, true);
//}
//}
//*/
//
//
//
///* DSMath.wpow
//function bpowi(uint a, uint n)
//internal pure
//returns (uint)
//{
//uint z = n % 2 != 0 ? a : BONE;
//
//for (n /= 2; n != 0; n /= 2) {
//a = bmul(a, a);
//
//if (n % 2 != 0) {
//z = bmul(z, a);
//}
//}
//return z;
//}
//*/
//
//// Power returns a the result of raising to a positive integer power
//func (d Dec) Power(power uint64) Dec {
//	if power == 0 {
//		return OneDec()
//	}
//	tmp := OneDec()
//	for i := power; i > 1; {
//		if i%2 == 0 {
//			i /= 2
//		} else {
//			tmp = tmp.Mul(d)
//			i = (i - 1) / 2
//		}
//		d = d.Mul(d)
//	}
//	return d.Mul(tmp)
//}
//
//
//func approxPoweri(base sdk.Dec, exp sdk.Uint) sdk.Dec {
//
//	if exp.Mod(sdk.NewUint(2)).Equal(sdk.ZeroUint())
//		z := sdk.NewUintFromBigInt(sdk.OneDec())
//	sdk.N
//	else
//	z := base
//
//	base.TruncateInt(sdk.OneDec())
//	uint z = n % 2 != 0 ? a : 63;
//
//
//	for (n /= 2; n != 0; n /= 2) {
//		a = bmul(a, a);
//
//		if (n % 2 != 0) {
//			z = bmul(z, a);
//		}
//	}
//	return z;
//}
//
//func approxPower(base sdk.Dec, exp sdk.Dec) (sdk.Dec, error) {
//
//	if base.LTE(sdk.ZeroDec()) || exp.LTE(sdk.ZeroDec()) {
//		return sdk.Dec{}, errors.New("base and exp can't be less than equal zero")
//	}
//}
//
//
///*
//function bpowApprox(uint base, uint exp, uint precision)
//internal pure
//returns (uint)
//{
//// term 0:
//uint a     = exp;
//(uint x, bool xneg)  = bsubSign(base, BONE);
//uint term = BONE;
//uint sum   = term;
//bool negative = false;
//
//
//// term(k) = numer / denom
////         = (product(a - i - 1, i=1-->k) * x^k) / (k!)
//// each iteration, multiply previous term by (a-(k-1)) * x / k
//// continue until term is less than precision
//for (uint i = 1; term >= precision; i++) {
//uint bigK = i * BONE;
//(uint c, bool cneg) = bsubSign(a, bsub(bigK, BONE));
//term = bmul(term, bmul(c, x));
//term = bdiv(term, bigK);
//if (term == 0) break;
//
//if (xneg) negative = !negative;
//if (cneg) negative = !negative;
//if (negative) {
//sum = bsub(sum, term);
//} else {
//sum = badd(sum, term);
//}
//}
//
//return sum;
//}
//
//}
//
//*/
//
//func bpowApprox(dec sdk.Dec base, uint exp, uint precision)
//returns (sdk.Dec)
//{
//// term 0:
//uint a     = exp;
//(uint x, bool xneg)  = bsubSign(base, BONE);
//uint term = BONE;
//uint sum   = term;
//bool negative = false;
//
//
//// term(k) = numer / denom
////         = (product(a - i - 1, i=1-->k) * x^k) / (k!)
//// each iteration, multiply previous term by (a-(k-1)) * x / k
//// continue until term is less than precision
//for (uint i = 1; term >= precision; i++) {
//uint bigK = i * BONE;
//(uint c, bool cneg) = bsubSign(a, bsub(bigK, BONE));
//term = bmul(term, bmul(c, x));
//term = bdiv(term, bigK);
//if (term == 0) break;
//
//if (xneg) negative = !negative;
//if (cneg) negative = !negative;
//if (negative) {
//sum = bsub(sum, term);
//} else {
//sum = badd(sum, term);
//}
//}
//
//return sum;
//}
//
//}
