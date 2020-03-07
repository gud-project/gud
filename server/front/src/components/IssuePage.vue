<template>
	<div>
		<h1>#{{ this.$route.params.issue }} : {{ this.issue.title }}</h1>
		<router-link :to="this.issue.author">@{{ this.issue.author }}</router-link><br /><br />
		<div>
			<p v-for="paragraph in issue.content.split('\n')">{{ paragraph }}</p>
		</div>
	</div>
</template>

<script>
	export default {
		name: "IssuePage",
		props: {
			pr: Boolean,
		},
		data() {
			return {
				issue: {
					title: null,
					author: null,
					content: null,
				},
			}
		},
		async created() {
			const { user, project, issue } = this.$route.params
			const res = await fetch(
				`/api/v1/user/${user}/project/${project}/${this.props.pr ? 'pr' : 'issue'}/${issue}`)
			if (res.ok) {
				this.issue = await res.json()
			} else {
				console.error(res.statusText)
			}
		}
	}
</script>

<style scoped>

</style>
