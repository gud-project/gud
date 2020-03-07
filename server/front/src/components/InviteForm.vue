<template>
	<form @submit="invite">
		<label><input type="search" v-model="name" placeholder="username" required /></label>
		<input type="submit" value="Invite" />
	</form>
</template>

<script>
	export default {
		name: "InviteForm",
		data() {
			return {
				name: null,
			}
		},
		methods: {
			async invite(e) {
				e.preventDefault()
				
				const { user, project } = this.$route.params
				const res = await fetch(`/user/${user}/project/${project}/invite`, {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ name: this.name })
				})
				
				if (res.ok) {
					this.name = ""
				} else {
					console.error(res.statusText)
				}
			}
		}
	}
</script>

<style scoped>

</style>
