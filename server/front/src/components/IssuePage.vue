<template>
	<div class="container">
		<div class="jumbotron">
			<h1>#{{ $route.params.issue }} : {{ issue.title }}</h1>
			<label>
				<router-link :to="issue.author">@{{ issue.author }}</router-link>
			</label>
			<br /><br />
			<p>{{ new Date(issue.created).toDateString() }}</p>
			<div v-if="pr">{{ issue.from }} ⇒ {{ issue.to }}</div>
			<p class="content">{{ issue.content }}</p>
		</div>
	</div>
</template>

<!--
<h1>#{{ $route.params.issue }} : {{ issue.title }}</h1>
<router-link :to="issue.author">@{{ issue.author }}</router-link><br /><br />
<div v-if="pr">{{ issue.from }} ⇒ {{ issue.to }}</div>
<p class="content">{{ issue.content }}</p>
-->
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
					from: null,
					to: null,
					created: null,
				},
			}
		},
		async created() {
			const { user, project, issue } = this.$route.params
			this.issue = await this.$getData(
				`/api/v1/user/${user}/project/${project}/${this.pr ? "prs" : "issues"}/${issue}`)
		}
	}
</script>

<style scoped>
.content {
	white-space: pre-line;
}
</style>
