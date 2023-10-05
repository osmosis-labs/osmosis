import unittest
import math

import scripts.cl.common.sdk_dec as sdk_dec

class SDKTestVector:
  def __init__(self, a: str, b: str, expected_mul: str, expected_quo: str, expected_quo_up: str, expected_quo_trunc: str):
        self.a = sdk_dec.new(a)
        self.b = sdk_dec.new(b)
        self.expected_mul = sdk_dec.new(expected_mul)
        self.expected_quo = sdk_dec.new(expected_quo)
        self.expected_quo_up = sdk_dec.new(expected_quo_up)
        self.expected_quo_trunc = sdk_dec.new(expected_quo_trunc)

# These test vectors are taken from:
# https://github.com/osmosis-labs/cosmos-sdk/blob/5c0181df6e93b72e7dacde8e08d6a59d316c732f/types/decimal_test.go#L184
sdk_test_vectors = [
    SDKTestVector("0", "0", "0", "0", "0", "0"),
    SDKTestVector("1", "0", "0", "0", "0", "0"),
    SDKTestVector("0", "1", "0", "0", "0", "0"),
    SDKTestVector("0", "-1", "0", "0", "0", "0"),
    SDKTestVector("-1", "0", "0", "0", "0", "0"),
    SDKTestVector("1", "1", "1", "1", "1", "1"),
    SDKTestVector("1", "1", "1", "1", "1", "1"),
    SDKTestVector("1", "2", "2", "0.5", "0.5", "0.5"),
    SDKTestVector("3", "7", "21", "0.428571428571428571", "0.428571428571428572", "0.428571428571428571"),
    SDKTestVector("3", "7", "21", "0.428571428571428571", "0.428571428571428572", "0.428571428571428571"),
    SDKTestVector("2", "4", "8", "0.5", "0.5", "0.5"),
    SDKTestVector("100", "100", "10000", "1", "1", "1"),
    SDKTestVector("0.15", "0.15", "0.022500000000000000", "1", "1", "1"),
    SDKTestVector("0.3333", "0.0333", "0.01109889", "10.009009009009009009", "10.009009009009009010", "10.009009009009009009"),
]

class TestSdkDec(unittest.TestCase):
    def test_mul(self):
        for vector in sdk_test_vectors:
            print(f"mul: {vector.a} x {vector.b} = {vector.expected_mul}")
            actual = sdk_dec.mul(vector.a, vector.b)
            self.assertEqual(vector.expected_mul, actual)

    def test_quo(self):
        for vector in sdk_test_vectors:
            if vector.b == sdk_dec.zero:
                continue

            print(f"quo: {vector.a} x {vector.b} = {vector.expected_quo}")
            actual = sdk_dec.quo(vector.a, vector.b)
            self.assertEqual(vector.expected_quo, actual)

    def test_quo_up(self):
        for vector in sdk_test_vectors:
            if vector.b == sdk_dec.zero:
                continue

            print(f"quo_up: {vector.a} x {vector.b} = {vector.expected_quo_up}")
            actual = sdk_dec.quo_up(vector.a, vector.b)
            self.assertEqual(vector.expected_quo_up, actual)


if __name__ == '__main__':
    unittest.main()