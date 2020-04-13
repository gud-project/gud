<template>
  <div id="app">
    <div id="nav">
      <router-link to="/">Home</router-link> |
	  <span v-if="loggedIn">
          <router-link to="/">Projects</router-link> |
		  <a @click="logOut">LogOut</a>
	  </span>
      <span v-else>
	      <router-link to="/SignUp">SignUp</router-link> |
	      <router-link to="/login">Login</router-link>
      </span>
    </div>
    <app-view />
  </div>
</template>

<script>
	export default {
		name: "App",
		data() {
			return {
				loggedIn: false,
			}
		},
		async created() {
			this.loggedIn = await this.$isLoggedIn()
		},
		methods: {
			async logOut() {
				await fetch('/api/v1/logout', { method: "POST" })
				await this.$router.go(0)
			}
		}
	}
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
}

#nav {
  padding: 30px;
}

#nav a {
  font-weight: bold;
  color: #2c3e50;
}

#nav a.router-link-exact-active {
  color: #2d72c0;
}
</style>
