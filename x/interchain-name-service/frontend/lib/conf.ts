export const getRpcEndpoint = () => {
  if (!process.env.NEXT_PUBLIC_RPC_ENDPOINT) {
    throw Error(
      "`NEXT_PUBLIC_RPC_ENDPOINT` env variable not found, please set"
    );
  }

  return process.env.NEXT_PUBLIC_RPC_ENDPOINT;
};

export const getChainId = () => {
  if (!process.env.NEXT_PUBLIC_CHAIN_ID) {
    throw Error("`NEXT_PUBLIC_CHAIN_ID` env variable not found, please set");
  }

  return process.env.NEXT_PUBLIC_CHAIN_ID;
};
