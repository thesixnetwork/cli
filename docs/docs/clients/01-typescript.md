---
sidebar_position: 1
description: Information about the generated Typescript client code.
---

# Typescript code generation

The `ignite generate ts-client` command generates a Typescript client for your blockchain project.

## Client code generation

A TypeScript (TS) client is automatically generated for your blockchain for custom and standard Cosmos SDK modules.

To enable client code generation, add the `client` entries to `config.yml`:

```yaml
client:
  typescript:
    path: "ts-client"
```

A TS client is generated in the `ts-client` directory.

## Client code regeneration

By default, the filesystem is watched and the clients are regenerated automatically. Clients for standard Cosmos SDK modules are generated after you scaffold a blockchain.

To regenerate all clients for custom and standard Cosmos SDK modules, run this command:

```bash
ignite generate ts-client
```

## Preventing client code regeneration	

To prevent regenerating the client, remove the `client:typescript` property from `config.yml`.	

## Usage

The code generated in `ts-client` comes with a `package.json` file ready to publish which you can modify to suit your needs.

The client is based on a modular architecture where you can configure a client class to support the modules you need and instantiate it.

By default, the generated client exports a client class that includes all the Cosmos SDK, custom and 3rd party modules in use in your project.

To instantiate the client you need to provide environment information (endpoints and chain prefix) and an optional wallet (implementing the CosmJS OfflineSigner interface).

For example, to connect to a local chain instance running under the Ignite CLI defaults, using Keplr as a wallet:

```typescript
import { Client } from '<path-to-ts-client>';

const client = new Client({ 
		apiURL: "http://localhost:1317",
		rpcURL: "http://localhost:26657",
		prefix: "cosmos"
	},
	window.keplr.getOfflineSigner()
);
```

The resulting client instance contains namespaces for each module, each with a `query` and `tx` namespace containing the module's relevant querying and transacting methods with full type and auto-completion support. 

e.g.

```typescript
const balances = await client.CosmosBankV1Beta1.query.queryAllBalances('cosmos1qqqsyqcyq5rqwzqfys8f67');
```

And for transactions:

```typescript
const tx_result = await client.CosmosBankV1Beta1.tx.sendMsgSend(
	{ 
		value: {
			amount: [
				{
					amount: '200',
					denom: 'token',
				},
			],
			fromAddress: 'cosmos1qqqsyqcyq5rqwzqfys8f67',
			toAddress: 'cosmos1qqqsyqcyq5rqwzqfys8f67'
		},
		fee,
		memo
	}
);
```

If you prefer, you can construct a lighter client using only the modules you are interested in by importing the generic client class and expanding it with the modules you need:

```typescript
import { IgniteClient } from '<path-to-ts-client>/client';
import { Module as CosmosBankV1Beta1 } from '<path-to-ts-client>/cosmos.bank.v1beta1'
import { Module as CosmosStakingV1Beta1 } from '<path-to-ts-client>/cosmos.staking.v1beta1'

const CustomClient = IgniteClient.plugin([CosmosBankV1Beta1, CosmosStakingV1Beta1]);

const client = new CustomClient({ 
		apiURL: "http://localhost:1317",
		rpcURL: "http://localhost:26657",
		prefix: "cosmos"
	},
	window.keplr.getOfflineSigner()
);
```

You can also construct TX messages separately and send them in a single TX using a global signing client like so:

```typescript
const msg1 = await client.CosmosBankV1Beta1.tx.msgSend(
	{ 
		value: {
			amount: [
				{
					amount: '200',
					denom: 'token',
				},
			],
			fromAddress: 'cosmos1qqqsyqcyq5rqwzqfys8f67',
			toAddress: 'cosmos1qqqsyqcyq5rqwzqfys8f67'
		}
	}
);
const msg2 = await client.CosmosBankV1Beta1.tx.msgSend(
	{ 
		value: {
			amount: [
				{
					amount: '200',
					denom: 'token',
				},
			],
			fromAddress: 'cosmos1qqqsyqcyq5rqwzqfys8f67',
			toAddress: 'cosmos1qqqsyqcyq5rqwzqfys8f67'
		},
	}
);
const tx_result = await client.signAndBroadcast([msg1,msg2], fee, memo);
```

Finally, for additional ease-of-use, apart from the modular client mentioned above, each generated module is usable on its own in a stripped-down way by exposing a separate txClient and queryClient.

e.g.

```typescript
import { queryClient } from '<path-to-ts-client>/cosmos.bank.v1beta1';

const client = queryClient({ addr: 'http://localhost:1317' });
const balances = await client.queryAllBalances('cosmos1qqqsyqcyq5rqwzqfys8f67');
```

and

```typescript
import { txClient } from '<path-to-ts-client>/cosmos.bank.v1beta1';

const client = txClient({
	signer: window.keplr.getOfflineSigner(),
	prefix: 'cosmos',
	addr: 'http://localhost:26657'
});

const tx_result = await client.sendMsgSend(
	{ 
		value: {
			amount: [
				{
					amount: '200',
					denom: 'token',
				},
			],
			fromAddress: 'cosmos1qqqsyqcyq5rqwzqfys8f67',
			toAddress: 'cosmos1qqqsyqcyq5rqwzqfys8f67'
		},
		fee,
		memo
	}
);
```