<template>
	<ul>
		<li v-for="project in projects" v-bind:key="project">
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
			const res = await fetch(`/api/v1/user/${this.$route.params.user}/projects`)
			if (res.ok) {
				this.projects = await res.json()
			} else {
				console.error(res.statusText)
			}
		},
	}
</script>

<style scoped>

</style>
