<template>
	<div>
		<div v-for="issue in issues">
			#{{ issue.id }}
			<router-link :to="`/${$route.params.user}/${$route.params.project}/issue/${issue.id}`">
				{{ issue.name }}
			</router-link>
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
