<template>
	<ul>
		<li v-for="name in projects">
			{{ name }}
		</li>
	</ul>
</template>

<script>
	export default {
		name: "ProjectsList",
		data() {
			return {
				projects: [],
			}
		},
		async created() {
			this.projects = await (await fetch('/api/v1/me/projects')).json()
		},
	}
</script>

<style scoped>

</style>
