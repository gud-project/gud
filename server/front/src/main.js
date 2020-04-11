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
			if (res.ok) {
				return await res.json()
			}
			this.$_error(Error, { error: res.status !== 404 && ((await res.json()).error || res.statusText) })
		},
	},
	computed: {
		async $loggedIn() {
			return (await fetch("/api/v1/me", { method: "HEAD" })).ok
		},
	},
})

new Vue({
	router,
	render: function (h) { return h(App) }
}).$mount("#app")
