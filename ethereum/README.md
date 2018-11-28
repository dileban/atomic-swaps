# Ethereum

## Contracts

**Migrations.sol**

This contract comes with every Truffle project and is not generally
edited. The contract is used to track the status of deployed contracts
on-chain and manage upgrades. By using Migrations only changed contracts
are deployed in subsequent deployments.

**GoldToken.sol**

A fictious fungible asset-backed token based on the
[ERC20](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-20.md)
standard. A single GoldToken is meant to represent an IOU
collateralized by gram of gold from a "trusted" institution. See
contract for token definition.

**HTLC.sol**

Defines an interface for Hashed TimeLock Contracts (HTLCs), sometimes
called Hashed TimeLock Agreements (HTLAs).

**AtomicSwap.sol**

An implementation of the HTLC interface enabling two parties to
transfer ownership of assets between them. Both parties will need to
agree on the terms of the exchange a priori. This includes valuation
of the respective assets and expiry time for the swap.

This contract is designed for swapping tokens on the same Ethereum
network instance (such as the Ethereum mainnet) or across different
network protocols (such as Ethereum and Hyperledger Fabric).

## Development

This project is based on the [Truffle
Framework](https://truffleframework.com/) and uses the ERC20
implementation from [OpenZeppelin](https://openzeppelin.org/).

Install dependencies (OpenZeppelin):

```
npm install
```

Compile contracts (see build directory):

```
truffle compile
```

## Deployment

There are a number of deployments options for testing purposes:

1. Truffle's built-in blockchain ([Truffle
   Develop](https://truffleframework.com/docs/truffle/getting-started/using-truffle-develop-and-the-console)).
2. [Ganache](https://truffleframework.com/ganache), a desktop
   application for launching a personal blockchain.
3. One of Ethereum's many testnets
   ([Ropsten](https://ropsten.etherscan.io/),
   [Kovan](https://kovan.etherscan.io/),
   [Rinkeby](https://rinkeby.etherscan.io/),
   [Sokol](https://sokol-explorer.poa.network/)).
4. The Ethereum mainnet.

Run `truffle develop` to use the console to deploy contracts within
truffle.

Alternatively, the deployment script `truffle.js` provides
configurations for deploying to both Ganache and Rinkeby.


### Ganache

Assumuing Ganache is running locally, to deploy to the default `development` network:

```
truffle migrate
```

### Rinkeby Testnet

[Rinkeby](https://www.rinkeby.io) is a Proof-of-Authority
testnet. Deploying will require Ether on the testnet. Use the
[faucet](https://www.rinkeby.io/#faucet) to obtain free Ether.





