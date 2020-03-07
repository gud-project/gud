import Vue from "vue"
import VueRouter from "vue-router"
import Home from "../views/Home.vue"

Vue.use(VueRouter)

const router = new VueRouter({
	routes: [
		{
			path: "/",
			name: "Home",
			component: Home,
		},
		{
			path: "/login",
			name: "Login",
			component() {
				return import(/* webpackChunkName: "login" */ "../views/Login.vue")
			},
		},
		{
			path: "/:user",
			name: "Projects",
			component() {
				return import(/* webpackChunkName: "user" */ "../views/User.vue")
			},
		},
		{
			path: "/:user/:project",
			name: "Project",
			component() {
				return import(/* webpackChunkName: "project" */ "../views/Project.vue")
			}
		},
	],
})

export default router
