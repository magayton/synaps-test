#!/bin/bash

ACCOUNT_NAME="payment"
ADDRESS="0xCAFE" # TO CHANGE

SCRIPT_DIR=$(dirname "$(realpath "$0")")

CONTRACT_DIR=$(realpath "$SCRIPT_DIR/..")

# cd to contract project directory (so that we can run the script from everywhere)
cd "$CONTRACT_DIR" || { echo "Erreur : Impossible de se déplacer dans le répertoire du contrat"; exit 1; }

echo "Compiling smart contract in Move"
aptos move compile --named-addresses "$ACCOUNT_NAME"="$ADDRESS"

# Verify compilation
if [ $? -ne 0 ]; then
  echo "Error while compiling smart contract."
  exit 1
fi

echo "Compilation success"

echo "Create object and publish package"
yes | aptos move create-object-and-publish-package --address-name "$ACCOUNT_NAME" --named-addresses "$ACCOUNT_NAME"="$ADDRESS"

# Verify object and publish success
if [ $? -ne 0 ]; then
  echo "Error while creating object and publishing"
  exit 1
fi

echo "Script success"
