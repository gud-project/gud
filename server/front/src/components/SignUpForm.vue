<template>
	<form @submit="signUp">
		<p style="color: red" v-for="error in errors">{{ error }}</p>
		<p><label>
			<input type="text" placeholder="Username" pattern="[a-zA-Z0-9_-]+" v-model="info.username" required />
		</label></p>
		<p><label>
			<input type="email" placeholder="E-mail" v-model="info.email" required />
		</label></p>
		<p><label>
			<input type="password" placeholder="Password" v-model="info.password" required />
		</label></p>
		<p><label>
			<input type="password" placeholder="Password (again)" v-model="passwordAgain" required />
		</label></p>
		
		<input type="submit" value="Sign Up" />
	</form>
</template>

<script>
	const PASS_MIN = 8
	
	export default {
		name: "SignUpForm",
		data() {
			return {
				info: {
					username: null,
					email: null,
					password: null,
				},
				passwordAgain: null,
				errors: [],
			}
		},
		methods: {
			async signUp(e) {
				e.preventDefault()
				
				if (this.info.password.length < PASS_MIN) {
					console.log("bad password")
				} else if (this.passwordAgain !== this.info.password) {
					console.log("bad repeat password")
				} else {
					const res = await fetch("/api/v1/signup", {
						method: "POST",
						headers: { 'Content-Type': 'application/json' },
						body: JSON.stringify(this.info)
					})
					
					if (res.ok) {
						await this.$router.push("/login")
					} else {
						this.errors = (await res.json()).errors
					}
				}
			},
		},
	}
</script>

<style scoped>

</style>
