<template>
	<div class="container">
		<div class="jumbotron">
			<h1>{{ $route.params.user }}</h1>
			<h2>{{ $route.params.project }}</h2>
			<br />
            <InviteForm />
			<br />
			<div class="container">
				<div class="container">
					<IssuesList category="issue" />
				</div>

				<div class="container">
					<IssuesList category="pr" />
				</div>
			</div>
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
