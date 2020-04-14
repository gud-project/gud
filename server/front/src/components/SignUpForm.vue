<template>
	
	<div class="container">
	<form @submit="signUp" class="jumbotron" style="margin:2.5em">
		<p style="color: red" v-for="error in errors" v-bind:key="error">{{ error }}</p>
		
		<p><label>
			<input type="text" class="form-control" placeholder="Username" pattern="[a-zA-Z0-9_-]+" v-model="info.username" required />
		</label></p>
		<p><label>
			<input type="email" class="form-control" placeholder="E-mail" v-model="info.email" required />
		</label></p>
		<p><label>
			<input type="password" class="form-control" placeholder="Password" v-model="info.password" required />
		</label></p>
		<p><label>
			<input type="password" class="form-control" placeholder="Password (again)" v-model="passwordAgain" required />
		</label></p>

		<input class="btn btn-primary btn-lg" type="submit" value="Sign Up" />
	</form>
	</div>

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
				
				this.errors = []
				if (this.info.password.length < PASS_MIN) {
					this.errors.push("password must be at least 8 characters long")
				}
				if (this.passwordAgain !== this.info.password) {
					this.errors.push("second password is different than first password")
				}
				if (this.errors.length > 0) {
					return false
				}
				
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
			},
		},
	}
</script>

<style scoped>

</style>
