# Synaps Technical Test
The repo is composed of two folders:
- aptos_contract: contains the move smart contracts
- golang_backend: contains the go backend

## Aptos_contract
The smart contracts are written in Move and designed for Aptos blockchain.
There are two smart contracts:
- soulbound: smart contract that mints soulbound tokens
- payment: smart contract that can receive payments in Aptos (APT) tokens

### Directories content 
- sources : folder containing one file with the smart contract source code
- Move.toml : file containing the dependencies of the smart contract
- scripts : folder containing one (bash) file to deploy the smart contract 

### How to build, test and deploy smart contracts
First of all you need the aptos-cli installed on your machine. You can find the installation guide [here](https://aptos.dev/en/build/cli).

Then you can build and test: 
```
cd aptos_contract/the-contract-you-want
aptos move compile --dev 
aptos move test 
```
To see test coverage you can do as follow (MODULE-NAME is payment or soulbound):
```
aptos move test --coverage --dev
(Optionnal) aptos move coverage source --module <MODULE-NAME> --dev
```
To deploy the smart contract you need to do the following __BEFORE__ executing the script:
- Create your account config for the devnet 
```
aptos init
```
Once done you should have a .aptos folder in the contract directory.
- Retrieve the account address in .aptos/config.yaml and put it in the deploy.sh script
```
ADDRESS="0xCAFE" # TO CHANGE
```
You can now deploy the smart contract with the following command:
```
./scripts/deploy.sh
```
The contract is now deployed and the CLI will return its address.

### Interacting with the contract through the CLI
If you want to interact with the smart contract through the CLI you can do as follow:
```
aptos move run --function-id CONTRACT_ADDRESS::MODULE::ENTRY_POINT --args ARGS
```
Here is an example for the payment contract (It sends 0.0000002 APT to the contract):
```
aptos move run --function-id 0x91aa6da54d1c187d1b9bfc9c89a5ce16c5d428e707e948c91ac8e1f3587b206b::payment::receive_payment --args u64:20
```
### Interacting with the contract through aptos explorer 
You can interact with the aptos explorer [here](https://explorer.aptoslabs.com/?network=devnet)

For example you can see the [payment contract](https://explorer.aptoslabs.com/account/0x91aa6da54d1c187d1b9bfc9c89a5ce16c5d428e707e948c91ac8e1f3587b206b) and interact with the Run / View tabs

## Golang_backend
The golang backend is made to interact with the soulboud smart contracts through an API endpoint to mint and to monitor the incoming payments on the payment smart contract

### Directories content 
- Makefile : A makefile to install dependecies, build the backend and test it
- .env : I purposely left this file here so that you can easily configure and change the backend configuration (but it should not be committed)
- go.mod and go.sum : The go modules used by the backend
- config/config.go : The configuration file to load .env file
- handler/services.go : The file containing the logic of the backend (minting, monitoring payments)
- main.go : The main file of the backend
- tests/mint_test.go and tests/payment_test.go : Tests of the backend

### How to build, test and deploy the golang backend
First of all you need the go installed on your machine. You can find the installation guide [here](https://go.dev/doc/install).

Then you can install the dependencies with the following command:
```
cd golang_backend
make install
```
Before testing the backend, you need to configure the .env file: 
```
PRIVATE_KEY="0x..."
MINT_CONTRACT_ADDRESS="0x..."
PAYMENT_CONTRACT_ADDRESS="0x..."
```
You can now build and test the backend with the following command:
```
make build
make test
```
Once built you can run the backend with the following command:
```
./aptos/services
```

### Interacting with the backend
You can use the mint endpoint to mint a soulbound token. 
```
curl -X POST http://localhost:8080/mint -H "Content-Type: application/json" -d '{"name": "NAME", "description": "Description", "base_uri": "URI", "soulbound_to": "ADDRESS"}'
```

You should also be able to monitor the payments on the payment contract while the backend is running.

## Helper 
If you do not want to deploy the smart contracts here are addresses of the deployed contracts on devnet:
- soulbound: 0xf95ab2fbc14e917b7124466a6df99544eb3b1173433a16f1cdbe33fd51c8d3ee
- payment: 0x91aa6da54d1c187d1b9bfc9c89a5ce16c5d428e707e948c91ac8e1f3587b206b