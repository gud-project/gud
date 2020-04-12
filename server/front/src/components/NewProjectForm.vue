<template>
	<form @submit="create">
		<label><input v-model="info.name" placeholder="Name" pattern="[a-zA-Z0-9_-]+" required /></label>
		<input type="submit" value="Create New Project" />
	</form>
</template>

<script>
	export default {
		name: "NewProjectForm",
        data() {
            return {
            	info: {
            		name: null,
	            },
            }
        },
		methods: {
			async create(e) {
				e.preventDefault()
				
				const res = await fetch("/api/v1/projects/create", {
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify(this.info),
				})
				if (res.ok) {
					this.$router.go(0)
				} else {
					alert((await res.json()).error || res.statusText)
				}
			},
		},
		
	}
</script>

<style scoped>

</style>
