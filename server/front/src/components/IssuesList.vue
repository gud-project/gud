<template>
	<div>
		<div v-for="issue in issues">
			#{{ issue.id }} <b>{{ issue.name }}</b>
			<br />
			<router-link :to="issue.author">@{{ issue.author }}</router-link>
			<br /><br />
		</div>
	</div>
</template>

<script>
	export default {
		name: "IssuesList",
		data() {
			return {
				issues: [],
			}
		},
		async created() {
			const { user, project } = this.$route.params
			this.issues = await (await fetch(`/api/v1/user/${user}/project/${project}/issues`)).json()
		},
	}
</script>

<style scoped>

</style>
