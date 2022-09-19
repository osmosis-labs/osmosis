import useSWR from "swr";
import { getAddress, getClient, getSigningClient } from "../lib/client";
import { getContractAddr } from "../lib/state";

export const getCount = async () => {
  const client = await getClient();
  return await client.queryContractSmart(getContractAddr(), { get_count: {} });
};

export const increase = async () => {
  const client = await getSigningClient();
  return await client.execute(
    await getAddress(),
    getContractAddr(),
    { increment: {} },
    "auto"
  );
};

export const useCount = () => {
  const { data, error, mutate } = useSWR("/counter/count", getCount);
  return {
    count: data?.count,
    error,
    increase: () => mutate(increase),
  };
};
