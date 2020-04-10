<template>
	<div>
		<h1>{{ $route.params.user }}</h1>
		<NewProjectForm />
		<ProjectsList />
	</div>
</template>

<script>
	import ProjectsList from "../components/ProjectsList"
	import NewProjectForm from "../components/NewProjectForm"
	export default {
		name: "Projects",
		components: { NewProjectForm, ProjectsList },
	}
</script>

<style scoped>

</style>
