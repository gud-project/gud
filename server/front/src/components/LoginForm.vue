<template>
	<form @submit="login">
		<p style="color: red" v-if="error">{{ error }}</p>
		<p><label>
			<input type="text" placeholder="Username" v-model="info.username" required />
		</label></p>
		<p><label>
			<input type="password" placeholder="Password" v-model="info.password" required />
		</label></p>
		<p><label>
			<input type="checkbox" v-model="info.remember" />
			Remember?
		</label></p>
		
		<input type="submit" value="Login" />
	</form>
</template>

<script>
	export default {
		name: "LoginForm",
		data() {
			return {
				info: {
					username: null,
					password: null,
					remember: false,
				},
				error: null,
			}
		},
		methods: {
			async login(e) {
				e.preventDefault()
				
				const res = await fetch("/api/v1/login", {
					method: "POST",
					credentials: "same-origin",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify(this.info),
				})
				
				if (res.ok) {
					await this.$router.push("/")
				} else {
					this.error = (await res.json()).error
				}
			}
		}
	}
</script>

<style scoped>

</style>
