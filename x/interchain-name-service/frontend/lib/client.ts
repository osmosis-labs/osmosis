import { CosmWasmClient, GasPrice, setupWebKeplr } from "cosmwasm";
import { getChainId, getRpcEndpoint } from "./conf";

export const getClient = () => CosmWasmClient.connect(getRpcEndpoint());
export const getAddress = async () => {
  if (!window.keplr) {
    throw Error("Unable to detect Keplr");
  }
  const chainId = getChainId();
  const offlineSigner = window.keplr.getOfflineSigner(chainId);
  const accounts = await offlineSigner.getAccounts();
  const address = accounts[0]?.address;

  if (!address) {
    throw Error("Signer address not found, please check your keplr wallet");
  }

  return address;
};

export const getSigningClient = () =>
  setupWebKeplr({
    rpcEndpoint: getRpcEndpoint(),
    prefix: "osmo",
    // Gas price has no effect if `preferNoSetFee` is set to false.
    // ref: https://docs.keplr.app/api/#interaction-options
    // but need to define since it's required by `SigningCosmWasmClient.execute`
    // when set fee to "auto"
    gasPrice: GasPrice.fromString("0uosmo"),
    chainId: getChainId(),
  });
