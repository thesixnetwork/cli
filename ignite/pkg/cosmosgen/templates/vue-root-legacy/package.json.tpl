{
  "name": "{{ .PackageNS }}-vuex-legacy",
  "version": "1.0.0",
  "description": "Legacy Vuex stores for {{ .PackageNS }} blockchain (v0.23.0 style)",
  "main": "index.js",
  "scripts": {
    "dev": "npm run build && npm run test",
    "build": "vue-cli-service build --target lib --name {{ .PackageNS }}-vuex src/index.js",
    "test": "echo \"No tests specified\""
  },
  "dependencies": {
    "@starport/vue": "^0.2.0",
    "@starport/vuex": "^0.2.0",
    "vue": "^2.6.14",
    "vuex": "^3.6.2",
    "axios": "^0.21.0"
  },
  "devDependencies": {
    "@vue/cli-service": "^4.5.0",
    "vue-template-compiler": "^2.6.14"
  },
  "repository": {
    "type": "git",
    "url": ""
  },
  "keywords": ["cosmos", "blockchain", "vuex", "legacy"],
  "author": "",
  "license": "MIT"
}
