import Vue from "vue"
import VueRouter from "vue-router"
import Home from "../views/Home.vue"

Vue.use(VueRouter)

const router = new VueRouter({
	routes: [
		{
			path: "/",
			name: "Home",
			component: Home
		},
		{
			path: "/login",
			name: "Login",
			component: function () {
				return import(/* webpackChunkName: "login" */ "../views/Login.vue")
			},
		},
	],
})

export default router
