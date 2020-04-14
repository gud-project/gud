<template>
	<div class="container">
		<div class="jumbotron">
			<h1>{{ $route.params.user }}</h1>
			<h2>{{ $route.params.project }}</h2>
			<br />
            <InviteForm />
			<br />
			<div class="container">
				<div class="container-fluid">
					<IssuesList category="issue" />
				</div>

				<div class="container-fluid">
					<PrsList category="pr" />
				</div>
			</div>
		</div>
	</div>
</template>

<script>
	import IssuesList from "../components/IssuesList"
	import InviteForm from "../components/InviteForm"
	import PrsList from "../components/PrsList"
	export default {
		name: "Project",
		components: { InviteForm, IssuesList, PrsList },
	}
</script>

<style scoped>
.container {
	display: flex;
	justify-content: space-evenly;
}
</style>
