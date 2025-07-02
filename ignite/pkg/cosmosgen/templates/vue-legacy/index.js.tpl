// Legacy Vuex store for {{ .Module.Pkg.Name }} module (v0.23.0 style)
import { txClient, queryClient, MissingWalletError, registry } from '../common'

export default {
  namespaced: true,
  
  state() {
    return {
      {{ range .Module.HTTPQueries }}{{ .Name }}: {},
      {{ end }}
      _Structure: {
        {{ range .Module.Types }}{{ .Name }}: {},
        {{ end }}
      },
      _Registry: registry,
      _Subscriptions: new Set(),
    }
  },

  getters: {
    {{ range .Module.HTTPQueries }}get{{ capitalCase .Name }}: (state) => (params = { params: {} }) => {
      if (!params.query) {
        params.query = {}
      }
      return state.{{ .Name }}[JSON.stringify(params)] ?? {}
    },
    {{ end }}
    getTypeStructure: (state) => (type) => {
      return state._Structure[type].fields
    },
    getRegistry: (state) => {
      return state._Registry
    }
  },

  mutations: {
    RESET_STATE(state) {
      Object.assign(state, getDefaultState())
    },
    {{ range .Module.HTTPQueries }}QUERY(state, { query, key, value }) {
      state[query][JSON.stringify(key)] = value
    },
    {{ end }}
    SUBSCRIBE(state, subscription) {
      state._Subscriptions.add(JSON.stringify(subscription))
    },
    UNSUBSCRIBE(state, subscription) {
      state._Subscriptions.delete(JSON.stringify(subscription))
    }
  },

  actions: {
    init({ dispatch, rootGetters }) {
      console.log('Initialized {{ .Module.Pkg.Name }} module store')
      if (rootGetters['common/env/client']) {
        rootGetters['common/env/client'].on('newblock', () => {
          dispatch('StoreUpdate')
        })
      }
    },
    
    resetState({ commit }) {
      commit('RESET_STATE')
    },

    unsubscribe({ commit }, subscription) {
      commit('UNSUBSCRIBE', subscription)
    },

    async StoreUpdate({ state, dispatch }) {
      state._Subscriptions.forEach(async (subscription) => {
        try {
          const sub = JSON.parse(subscription)
          await dispatch(sub.action, sub.payload)
        } catch (e) {
          throw new Error('Subscriptions: ' + e.message)
        }
      })
    },

    {{ range .Module.HTTPQueries }}async Query{{ capitalCase .Name }}({ commit, rootGetters, getters }, { options: { subscribe, all } = { subscribe: false, all: false }, params, query = null }) {
      try {
        const key = params ?? {}
        const queryClient = await queryClient(rootGetters)
        let value = (await queryClient.query{{ capitalCase .Module.Pkg.Name }}.{{ camelCase .Name }}(key)).data
        
        commit('QUERY', { query: '{{ .Name }}', key: { params: { ...key }, query }, value })
        
        if (subscribe) {
          commit('SUBSCRIBE', { action: 'Query{{ capitalCase .Name }}', payload: { options: { all }, params: { ...key }, query } })
        }
        
        return getters['get{{ capitalCase .Name }}']({ params: { ...key }, query }) ?? {}
      } catch (e) {
        throw new Error('Query{{ capitalCase .Name }}: ' + e.message)
      }
    },
    
    {{ end }}
    {{ range .Module.Messages }}async Msg{{ capitalCase .Name }}({ rootGetters }, { value }) {
      try {
        const txClient = await txClient(rootGetters)
        const msg = await txClient.msg{{ capitalCase .Module.Pkg.Name }}.{{ camelCase .Name }}({ value })
        return await txClient.signAndBroadcast([msg], { fee: { amount: [], gas: "200000" }, memo: "" })
      } catch (e) {
        if (e == MissingWalletError) {
          throw new Error('TxClient:Msg{{ capitalCase .Name }}:Init: Unable to sign and broadcast transaction. ' + e.message)
        } else {
          throw new Error('TxClient:Msg{{ capitalCase .Name }}:Send: Could not broadcast Tx: ' + e.message)
        }
      }
    },
    
    {{ end }}
  }
}

function getDefaultState() {
  return {
    {{ range .Module.HTTPQueries }}{{ .Name }}: {},
    {{ end }}
    _Structure: {
      {{ range .Module.Types }}{{ .Name }}: {},
      {{ end }}
    },
    _Registry: registry,
    _Subscriptions: new Set(),
  }
}
