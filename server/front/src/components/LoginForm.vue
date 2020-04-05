<template>
	<div class="container">
		<form @submit="login" class="jumbotron" style="margin:2.5em">
			<p style="color: red" v-if="error">{{ error }}</p>
			<p><label>
				<input type="text" placeholder="Username" class="form-control" v-model="info.username" required />
			</label></p>
			<p><label>
				<input type="password" placeholder="Password" class="form-control" v-model="info.password" required />
			</label></p>
			<p><label>
				<div class="custom-control custom-checkbox">
					<input type="checkbox" class="custom-control-input" id="customCheck1">
					<label class="custom-control-label" for="customCheck1">Remember?</label>
				</div>
			</label></p>
			<input class="btn btn-primary btn-lg" type="submit" value="Login" />
		</form>
	</div>
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
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify(this.info),
				})
				
				if (res.ok) {
					await this.$router.push(`/${this.info.username}`)
				} else {
					this.error = (await res.json()).error
				}
			}
		}
	}
</script>

<style scoped>

</style>
