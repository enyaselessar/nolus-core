#!/bin/bash
set -euox pipefail
# TBD supersede penultimate-genesis.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR"/internal/genesis.sh

cleanup() {
  cleanup_genesis_sh
  exit
}
trap cleanup INT TERM EXIT

__print_usage() {
    printf \
    "Usage: %s
    <$COMMAND_FULL_GEN>
    [-c|--chain_id <string>]
    [--currency <native_currency>]
    [--accounts <accounts_spec_json>]
    [--validator-node-urls-pubkeys <validator_node_urls_and_validator_pubkeys>]
    [--validator-accounts-dir <validator_accounts_dir>]
    [--validator-tokens <validators_initial_tokens>]
    [--validator-stake <tokens_validator_stakes>]
    [-o|--output <genesis_file_path>]" \
     "$1"
}

__verify_mandatory() {
  local value="$1"
  local description="$2"

  if [[ -z "$value" ]]; then
    echo >&2 "$description was not set"
    exit 1
  fi
}

COMMAND_FULL_GEN="full-gen"
CHAIN_ID=""
NATIVE_CURRENCY="unolus"
ACCOUNTS_SPEC=""
VAL_NODE_URLS_AND_VAL_PUBKEYS=""
VAL_ACCOUNTS_DIR="val-accounts"
VAL_TOKENS="1000000000""$NATIVE_CURRENCY"
VAL_STAKE="1000000""$NATIVE_CURRENCY"
OUTPUT_FILE=""

if [[ $# -lt 1 ]]; then
  echo "Missing command!"
  __print_usage "$0"
  exit 1
fi
COMMAND="$1"
shift

while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in

  -h | --help)
    __print_usage "$0"
    exit 0
    ;;

  -c | --chain-id)
    CHAIN_ID="$2"
    shift
    shift
    ;;

  --currency)
    NATIVE_CURRENCY="$2"
    shift
    shift
    ;;

  --accounts)
    ACCOUNTS_SPEC="$2"
    shift
    shift
    ;;

  --validator-node-urls-pubkeys)
    VAL_NODE_URLS_AND_VAL_PUBKEYS="$2"
    shift
    shift
    ;;

  --validator-accounts-dir)
    VAL_ACCOUNTS_DIR="$2"
    shift
    shift
    ;;

  --validator-tokens)
    VAL_TOKENS="$2"
    shift
    shift
    ;;

  --validator-stake)
    VAL_STAKE="$2"
    shift
    shift
    ;;

  -o | --output)
    OUTPUT_FILE="$2"
    shift
    shift
    ;;
  
  *)
    echo "unknown option '$key'"
    exit 1
    ;;

  esac
done

if [[ "$COMMAND" == "$COMMAND_FULL_GEN" ]]; then
  __verify_mandatory "$CHAIN_ID" "Nolus chain identifier"
  __verify_mandatory "$ACCOUNTS_SPEC" "Nolus genesis accounts spec"
  __verify_mandatory "$VAL_NODE_URLS_AND_VAL_PUBKEYS" "Validator URLs and validator public keys spec"
  __verify_mandatory "$OUTPUT_FILE" "Genesis output file"

  genesis_file=$(generate_genesis "$CHAIN_ID" "$NATIVE_CURRENCY" "$VAL_TOKENS" "$VAL_STAKE" \
                                  "$ACCOUNTS_SPEC" "$VAL_ACCOUNTS_DIR" \
                                  "$VAL_NODE_URLS_AND_VAL_PUBKEYS")
  mv "$genesis_file" "$OUTPUT_FILE"
# elif [[ "$COMMAND" == "$COMMAND_SETUP" ]]; then
#
else
  echo "Unknown command!"
  exit 1
fi