<template>
	<div class="container">
		<div class="jumbotron">
			<h1>#{{ $route.params.issue }} : {{ issue.title }}</h1>
			<label>
				<router-link :to="issue.author">@{{ issue.author }}</router-link>
			</label>
			<br />
			<p class="content">{{ issue.content }}</p>
			<form @submit="setStatus">
				<select v-model="status" class="custom-select" style="width:auto;">
					<option value="open">Open</option>
					<option value="in_progress">In Progress</option>
					<option value="done">Done</option>
					<option value="closed">Closed</option>
				</select>
				<input type="submit" value="Set Status" class="btn btn-success" :disabled="status === issue.status" />
			</form>
			<br /><br />
			<p>{{ new Date(issue.created).toDateString() }}</p>
		</div>
	</div>
</template>

<script>
	export default {
		name: "IssuePage",
		data() {
			return {
				issue: {
					title: null,
					author: null,
					content: null,
					status: null,
					created: null,
				},
				status: null,
			}
		},
		async created() {
			const { user, project, issue } = this.$route.params
			this.issue = await this.$getData(
				`/api/v1/user/${user}/project/${project}/issues/${issue}`)
			this.status = this.issue.status
		},
		methods: {
			async setStatus(e) {
				e.preventDefault()
				
				const { user, project, issue } = this.$route.params
				const res = await fetch(`/api/v1/user/${user}/project/${project}/issues/${issue}/update`, {
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify({ status: this.status })
				})
				
				if (res.ok) {
					this.issue.status = this.status
				} else {
					alert(res.status !== 404 && ((await res.json()).error || res.statusText))
				}
			},
		}
	}
</script>

<style scoped>
.content {
	white-space: pre-line;
	font-size: larger;
}
</style>
