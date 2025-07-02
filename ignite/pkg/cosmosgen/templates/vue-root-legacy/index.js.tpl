// Legacy JavaScript client for {{ .PackageNS }}
import { Client, registry, MissingWalletError } from '@starport/vue'

const Modules = {
  {{ range .Modules }}{{ .Pkg.Name }}: () => import('./{{ .Pkg.Name }}'),
  {{ end }}
}

export default {
  Modules,
  registry,
  MissingWalletError,
  
  install(Vue, options = {}) {
    const client = new Client(options.env || {}, options.wallet)
    
    Vue.prototype.$client = client
    Vue.prototype.$registry = registry
    
    // Register all modules
    Object.keys(Modules).forEach(async (moduleName) => {
      const module = await Modules[moduleName]()
      Vue.prototype.$client.registerModule(moduleName, module.default)
    })
    
    return client
  }
}
