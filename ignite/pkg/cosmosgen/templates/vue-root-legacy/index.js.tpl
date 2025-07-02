// Legacy Vuex stores index (v0.23.0 style)
import { Client } from '@starport/vuex'

// Import all module stores
{{ range .Modules }}import {{ .Pkg.Name }} from './{{ .Pkg.Name }}'
{{ end }}

const modules = {
  {{ range .Modules }}{{ .Pkg.Name }},
  {{ end }}
}

export default function initStores(store) {
  // Register all modules with the store
  Object.keys(modules).forEach(name => {
    store.registerModule(['{{ .PackageNS }}', name], modules[name])
  })
  
  // Initialize client for legacy behavior
  const client = new Client()
  
  return {
    modules,
    client
  }
}
