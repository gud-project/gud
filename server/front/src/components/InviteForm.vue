<template>
	<form @submit="invite">
		<label><input type="search" class="form-control" v-model="name" placeholder="username" required/></label>
		<input type="submit" class="btn btn-outline-primary" value="Invite" />
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
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify({ name: this.name })
				})
				
				if (res.ok) {
					this.name = ""
				} else {
					alert((await res.json()).error || res.statusText)
				}
			}
		}
	}
</script>

<style scoped>

</style>
