<template>
	<div>
		<h1>{{ $route.params.user }}/{{ $route.params.project }}</h1>
		<InviteForm />
		<div class="container">
			<IssuesList category="issue" />
			<IssuesList category="pr" />
		</div>
	</div>
</template>

<script>
	import IssuesList from "../components/IssuesList"
	import InviteForm from "../components/InviteForm"
	export default {
		name: "Project",
		components: { InviteForm, IssuesList },
	}
</script>

<style scoped>
.container {
	display: flex;
	justify-content: space-evenly;
}
</style>
