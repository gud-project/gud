<template>
	<div>
		<router-link :to="`/${$route.params.user}/${$route.params.project}/${category}/new`">
			New {{ category }}
		</router-link>
		<div v-for="issue in issues">
			#{{ issue.id }}
			<router-link :to="`/${$route.params.user}/${$route.params.project}/${category}/${issue.id}`">
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
		props: {
			category: {
				type: String,
				default: 'issue',
			},
		},
		data() {
			return {
				issues: [],
			}
		},
		async created() {
			const { user, project } = this.$route.params
			const res = await fetch(`/api/v1/user/${user}/project/${project}/${this.props.category}s`)
			
			if (res.ok) {
				this.issues = await res.json()
			} else {
				console.error(res.statusText)
			}
		},
	}
</script>

<style scoped>

</style>
