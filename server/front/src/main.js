import Vue from "vue"
import ErrorPage from "vue-error-page/src"
import App from "./App.vue"
import router from "./router"
import Error from "./views/Error"

Vue.config.productionTip = false
window.eventBus = new Vue()

Vue.use(ErrorPage)

Vue.mixin({
	methods: {
		async $getData(url) {
			const res = await fetch(url)
			console.log(res)
			if (res.ok) {
				return await res.json()
			}
			
			let error
			try {
				error = (await res.json()).error || res.statusText
			} catch {
				error = res.statusText
			}
			this.$_error(Error, { error })
		},
		async $getUser() {
			const res = await fetch("/api/v1/me")
			return res.ok ? (await res.json()).username : null
		},
	},
})

new Vue({
	router,
	render: function (h) { return h(App) }
}).$mount("#app")
