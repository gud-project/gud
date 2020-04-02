<template>
	<div>
		<h1>{{ $route.params.user }}</h1>
		<ProjectsList />
	</div>
</template>

<script>
	import ProjectsList from "../components/ProjectsList"
	export default {
		name: "Projects",
		components: { ProjectsList }
	}
</script>

<style scoped>

</style>
