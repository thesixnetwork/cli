// Legacy common helpers (v0.23.0 style)
import { Client } from '@starport/vuex'

export { MissingWalletError } from '@starport/vuex'

let _client
let _txClient  
let _queryClient

export function getClient() {
  if (!_client) {
    _client = new Client()
  }
  return _client
}

export async function txClient(rootGetters) {
  if (!_txClient) {
    const client = getClient()
    const signer = rootGetters['common/wallet/signer']
    const env = rootGetters['common/env/getEnv']
    
    _txClient = await client.initTxClient(signer, env)
  }
  return _txClient
}

export async function queryClient(rootGetters) {
  if (!_queryClient) {
    const client = getClient()
    const env = rootGetters['common/env/getEnv']
    
    _queryClient = await client.initQueryClient(env)
  }
  return _queryClient
}

// Legacy registry export
export const registry = []
