import { getChainId } from "../connectors/utils"

const chainID = getChainId()
export const useEtherscanUrl = () => {
  return chainID === 30 // Mainnet network ID.
    ? "https://explorer.rsk.co/"
    : "https://explorer.testnet.rsk.co"
}
