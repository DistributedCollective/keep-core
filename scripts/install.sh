#!/bin/bash
set -e

# Dafault inputs.
KEEP_ETHEREUM_PASSWORD_DEFAULT="password"
KEEP_CORE_PATH=$PWD
CONFIG_DIR_PATH_DEFAULT="$KEEP_CORE_PATH/configs"

if [[ -z "${DEST_NETWORK}" ]]; then
  echo "DEST_NETWORK env not set. Exiting"
  exit 1
else
  echo "Using network ${DEST_NETWORK}"
fi

if [[ -z "${INITIALIZE_OPERATORS}" ]]; then
  echo "INITIALIZE_OPERATORS env not set, Exiting"
  exit 1
fi

# Read user inputs.
read -p "Enter ethereum accounts password [$KEEP_ETHEREUM_PASSWORD_DEFAULT]: " ethereum_password
KEEP_ETHEREUM_PASSWORD=${ethereum_password:-$KEEP_ETHEREUM_PASSWORD_DEFAULT}

read -p "Enter dir path to keep-core client configs [$CONFIG_DIR_PATH_DEFAULT]: " config_dir_path
CONFIG_DIR_PATH=${config_dir_path:-$CONFIG_DIR_PATH_DEFAULT}

# Run script.
LOG_START='\n\e[1;36m' # new line + bold + color
LOG_END='\n\e[0m' # new line + reset color

printf "${LOG_START}Starting installation...${LOG_END}"
KEEP_CORE_CONFIG_DIR_PATH=$CONFIG_DIR_PATH
KEEP_CORE_SOL_PATH="$KEEP_CORE_PATH/solidity"

cd $KEEP_CORE_SOL_PATH

printf "${LOG_START}Installing NPM dependencies...${LOG_END}"
npm install

printf "${LOG_START}Unlocking ethereum accounts...${LOG_END}"
KEEP_ETHEREUM_PASSWORD=$KEEP_ETHEREUM_PASSWORD \
    npx truffle exec scripts/unlock-eth-accounts.js --network $DEST_NETWORK

printf "${LOG_START}Migrating contracts...${LOG_END}"
rm -rf build/
npx truffle compile
npx truffle migrate --reset --network $DEST_NETWORK

KEEP_CORE_SOL_ARTIFACTS_PATH="$KEEP_CORE_SOL_PATH/build/contracts"

if [[ $INITIALIZE_OPERATORS != 0 ]]; then 
  printf "${LOG_START}Initializing contracts...${LOG_END}"
  npx truffle exec scripts/delegate-tokens.js --network $DEST_NETWORK
fi

printf "${LOG_START}Updating keep-core client configs...${LOG_END}"
for CONFIG_FILE in $KEEP_CORE_CONFIG_DIR_PATH/*.toml
do
    KEEP_CORE_CONFIG_FILE_PATH=$CONFIG_FILE \
        npx truffle exec scripts/lcl-client-config.js --network $DEST_NETWORK
done

printf "${LOG_START}Building keep-core client...${LOG_END}"
cd $KEEP_CORE_PATH
go generate ./...
go build -a -o keep-core .
