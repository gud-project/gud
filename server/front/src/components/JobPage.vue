<template>
	<div>
		<h1>#{{ $route.params.job }} : {{ job.status }}</h1>
		<p class="logs">{{ job.content }}</p>
	</div>
</template>

<script>
	export default {
		name: "JobPage",
		data() {
			return {
				job: {
					status: null,
					logs: null,
				},
			}
		},
		async created() {
			const { user, project, job } = this.$route.params
			const res = await fetch(
				`/api/v1/user/${user}/project/${project}/job/${job}`)
			if (res.ok) {
				this.job = await res.json()
			} else {
				console.error(res.statusText)
			}
		}
	}
</script>

<style scoped>
.logs {
	white-space: pre-line;
	background-color: black;
	color: white;
}
</style>
