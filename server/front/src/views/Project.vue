<template>
	<div class="container">
		<IssuesList category="issue" />
		<IssuesList category="pr" />
	</div>
</template>

<script>
	import IssuesList from "../components/IssuesList"
	export default {
		name: "Project",
		components: { IssuesList },
	}
</script>

<style scoped>
.container {
	display: flex;
	justify-content: space-evenly;
}
</style>
