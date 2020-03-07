<template>
	<ul>
		<li v-for="project in projects">
			<router-link :to="`/${$route.params.user}/${project}`">{{ project }}</router-link>
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
			this.projects = await (await fetch(`/api/v1/user/${this.$route.params.user}/projects`)).json()
		},
	}
</script>

<style scoped>

</style>
