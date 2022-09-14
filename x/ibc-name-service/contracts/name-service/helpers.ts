/*
 * This is a set of helpers meant for use with @cosmjs/cli
 * With these you can easily use the cw20 contract without worrying about forming messages and parsing queries.
 *
 * Usage: npx @cosmjs/cli@^0.23 --init https://raw.githubusercontent.com/CosmWasm/cosmwasm-examples/main/nameservice/helpers.ts
 *
 * Create a client:
 *   const client = await useOptions(coralnetOptions).setup(password);
 *   await client.getAccount()
 *
 * Get the mnemonic:
 *   await useOptions(coralnetOptions).recoverMnemonic(password)
 *
 * If you want to use this code inside an app, you will need several imports from https://github.com/CosmWasm/cosmjs
 */

const path = require("path");

interface Options {
  readonly httpUrl: string
  readonly networkId: string
  readonly feeToken: string
  readonly gasPrice: GasPrice
  readonly bech32prefix: string
  readonly hdPath: HdPath
  readonly faucetUrl?: string
  readonly defaultKeyFile: string
  readonly gasLimits: Partial<GasLimits<CosmWasmFeeTable>> // only set the ones you want to override
}

const coralnetOptions: Options = {
  httpUrl: 'https://lcd.coralnet.cosmwasm.com',
  networkId: 'cosmwasm-coral',
  feeToken: 'ucosm',
  gasPrice:  GasPrice.fromString("0.025ucosm"),
  bech32prefix: 'cosmos',
  faucetToken: 'SHELL',
  faucetUrl: 'https://faucet.coralnet.cosmwasm.com/credit',
  hdPath: makeCosmoshubPath(0),
  defaultKeyFile: path.join(process.env.HOME, ".coral.key"),
  gasLimits: {
    upload: 1500000,
    init: 600000,
    register:800000,
    transfer: 80000,
  },
}

interface Network {
  setup: (password: string, filename?: string) => Promise<SigningCosmWasmClient>
  recoverMnemonic: (password: string, filename?: string) => Promise<string>
}

const useOptions = (options: Options): Network => {

  const loadOrCreateWallet = async (options: Options, filename: string, password: string): Promise<Secp256k1HdWallet> => {
    let encrypted: string;
    try {
      encrypted = fs.readFileSync(filename, 'utf8');
    } catch (err) {
      // generate if no file exists
      const wallet = await Secp256k1HdWallet.generate(12, options.hdPath, options.bech32prefix);
      const encrypted = await wallet.serialize(password);
      fs.writeFileSync(filename, encrypted, 'utf8');
      return wallet;
    }
    // otherwise, decrypt the file (we cannot put deserialize inside try or it will over-write on a bad password)
    const wallet = await Secp256k1HdWallet.deserialize(encrypted, password);
    return wallet;
  };

  const connect = async (
    wallet: Secp256k1HdWallet,
    options: Options
  ): Promise<SigningCosmWasmClient> => {
    const [{ address }] = await wallet.getAccounts();

    const client = new SigningCosmWasmClient(
      options.httpUrl,
      address,
      wallet,
      coralnetOptions.gasPrice,
      coralnetOptions.gasLimits,
    );
    return client;
  };

  const hitFaucet = async (
    faucetUrl: string,
    address: string,
    denom: string
  ): Promise<void> => {
    await axios.post(faucetUrl, { denom, address });
  }

  const setup = async (password: string, filename?: string): Promise<SigningCosmWasmClient> => {
    const keyfile = filename || options.defaultKeyFile;
    const wallet = await loadOrCreateWallet(coralnetOptions, keyfile, password);
    const client = await connect(wallet, coralnetOptions);

    // ensure we have some tokens
    if (options.faucetUrl) {
      const account = await client.getAccount();
      if (!account) {
        console.log(`Getting ${options.feeToken} from faucet`);
        await hitFaucet(options.faucetUrl, client.senderAddress, options.feeToken);
      }
    }

    return client;
  }

  const recoverMnemonic = async (password: string, filename?: string): Promise<string> => {
    const keyfile = filename || options.defaultKeyFile;
    const wallet = await loadOrCreateWallet(coralnetOptions, keyfile, password);
    return wallet.mnemonic;
  }

  return {setup, recoverMnemonic};
}

interface Config {
  readonly purchase_price?: Coin
  readonly transfer_price?: Coin
}

interface ResolveRecordResponse {
  readonly address?: string
}

interface InitMsg {

  readonly purchase_price?: Coin
  readonly transfer_price?: Coin
}

interface NameServiceInstance {
  readonly contractAddress: string

  // queries
  record: (name: string) => Promise<ResolveRecordResponse>
  config: () => Promise<Config>

  // actions
  register: (name: string, amount: Coin[]) => Promise<any>
  transfer: (name: string, to: string, amount: Coin[]) => Promise<any>
}

interface NameServiceContract {
  upload: () => Promise<number>

  instantiate: (codeId: number, initMsg: InitMsg, label: string) => Promise<NameServiceInstance>

  use: (contractAddress: string) => NameServiceInstance
}

const NameService = (client: SigningCosmWasmClient): NameServiceContract => {
  const use = (contractAddress: string): NameServiceInstance => {
    const record = async (name: string): Promise<ResolveRecordResponse> => {
      return client.queryContractSmart(contractAddress, {resolve_record: { name }});
    };

    const config = async (): Promise<Config> => {
      return client.queryContractSmart(contractAddress, {config: { }});
    };

    const register = async (name: string, amount: Coin[]): Promise<any> => {
      const result = await client.execute(contractAddress, {register: { name }}, "", amount);
      return result.transactionHash;
    };

    const transfer = async (name: string, to: string, amount: Coin[]): Promise<any> => {
      const result = await client.execute(contractAddress, {transfer: { name, to }}, "", amount);
      return result.transactionHash;
    };

    return {
      contractAddress,
      record,
      config,
      register,
      transfer,
    };
  }

  const downloadWasm = async (url: string): Promise<Uint8Array> => {
    const r = await axios.get(url, { responseType: 'arraybuffer' })
    if (r.status !== 200) {
      throw new Error(`Download error: ${r.status}`)
    }
    return r.data
  }

  const upload = async (): Promise<number> => {
    const meta = {
      source: "https://github.com/CosmWasm/cosmwasm-examples/tree/nameservice-0.7.0/nameservice",
      builder: "cosmwasm/rust-optimizer:0.10.4"
    };
    const sourceUrl = "https://github.com/CosmWasm/cosmwasm-examples/releases/download/nameservice-0.7.0/contract.wasm";
    const wasm = await downloadWasm(sourceUrl);
    const result = await client.upload(wasm, meta);
    return result.codeId;
  }

  const instantiate = async (codeId: number, initMsg: Record<string, unknown>, label: string): Promise<NameServiceInstance> => {
    const result = await client.instantiate(codeId, initMsg, label, { memo: `Init ${label}`});
    return use(result.contractAddress);
  }

  return { upload, instantiate, use };
}

// Demo:
// const client = await useOptions(coralnetOptions).setup(PASSWORD);
// const { address } = await client.getAccount()
// const factory = NameService(client)
//
// const codeId = await factory.upload();
// codeId -> 12
// const initMsg = { purchase_price: { denom: "ushell", amount:"1000" }, transfer_price: { denom: "ushell", amount:"1000" }}
// const contract = await factory.instantiate(12, initMsg, "My Name Service")
// contract.contractAddress -> 'coral1267wq2zk22kt5juypdczw3k4wxhc4z47mug9fd'
//
// OR
//
// const contract = factory.use('coral1267wq2zk22kt5juypdczw3k4wxhc4z47mug9fd')
//
// const randomAddress = 'coral162d3zk45ufaqke5wgcd3kh336k6p3kwwkdj3ma'
//
// contract.config()
// contract.register("name", [{"denom": "ushell", amount: "4000" }])
// contract.record("name")
// contract.transfer("name", randomAddress, [{"denom": "ushell", amount: "2000" }])
//
