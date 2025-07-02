// Legacy common utilities for {{ .PackageNS }}
import { Client, registry } from '@starport/vue'

export { registry, MissingWalletError } from '@starport/vue'

let _client = null

export async function txClient(rootGetters) {
  if (!_client) {
    const { address } = rootGetters['common/wallet']
    const client = new Client(rootGetters['common/env'], rootGetters['common/wallet'])
    await client.connectSigner(address)
    _client = client
  }
  return _client
}

export async function queryClient(rootGetters) {
  if (!_client) {
    const client = new Client(rootGetters['common/env'])
    await client.initStargateClient()
    _client = client
  }
  return _client
}

// Reset client connection
export function resetClient() {
  _client = null
}
