import Vue from "vue"
import VueRouter from "vue-router"
import Home from "../views/Home.vue"
import Error from "../views/Error"

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
				return import(/* webpackChunkName: "login" */ "../views/Login")
			},
		},
		{
			path: "/signup",
			name: "signup",
			component() {
				return import(/* webpackChunkName: "signup" */ "../views/SignUp")
			},
		},
		{
			path: "/:user",
			name: "Projects",
			component() {
				return import(/* webpackChunkName: "user" */ "../views/User")
			},
		},
		{
			path: "/:user/:project",
			name: "Project",
			component() {
				return import(/* webpackChunkName: "project" */ "../views/Project")
			},
		},
		{
			path: "/:user/:project/issue/new",
			name: "NewIssue",
			component() {
				return import(/* webpackChunkName: "newIssue" */"../views/NewIssue")
			},
		},
		{
			path: "/:user/:project/pr/new",
			name: "NewPR",
			component() {
				return import(/* webpackChunkName: "newIssue" */"../views/NewPR")
			},
		},
		{
			path: "/:user/:project/issue/:issue",
			name: "Issue",
			component() {
				return import(/* webpackChunkName: "issue" */ "../views/Issue")
			},
		},
		{
			path: "/:user/:project/pr/:pr",
			name: "PR",
			component() {
				return import(/* webpackChunkName: "issue" */ "../views/PR")
			},
		},
		{
			path: "/:user/:project/job/:job",
			name: "Job",
			component() {
				return import(/* webpackChunkName: "job" */ "../views/Job")
			},
		},
		{
			path: "*",
			name: Error,
			component: Error,
		},
	],
})

export default router
