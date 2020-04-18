<template>
	<div class="container">
		<div class="jumbotron">
			<h1>#{{ $route.params.issue }} : {{ issue.title }}</h1>
			<label>
				<router-link :to="issue.author">@{{ issue.author }}</router-link>
			</label>
			<br /><br />
			<p>{{ new Date(issue.created).toDateString() }}</p>
			<p v-if="pr"><b>{{ issue.status }}</b> {{ issue.from }} ⇒ {{ issue.to }}</p>
			<p class="content">{{ issue.content }}</p>
			
			<div v-if="pr">
				<button @click="mergePr" :disabled="status !== 'open'">Merge</button>
				<button @click="closePr" :disabled="status !== 'open'">Close</button>
			</div>
			<form v-else @submit="setStatus">
				<select v-model="status">
					<option value="open">Open</option>
					<option value="in_progress">In Progress</option>
					<option value="done">Done</option>
					<option value="closed">Closed</option>
				</select>
				<input type="submit" value="Set Status" :disabled="status === issue.status" />
			</form>
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
					status: null,
					from: null,
					to: null,
					created: null,
				},
				status: null,
			}
		},
		async created() {
			const { user, project, issue } = this.$route.params
			this.issue = await this.$getData(
				`/api/v1/user/${user}/project/${project}/${this.pr ? "prs" : "issues"}/${issue}`)
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
			async mergePr() {
				const { user, project, issue } = this.$route.params
				const res = await fetch(`/api/v1/user/${user}/project/${project}/prs/${issue}/merge`, {
					method: "POST",
				})
				
				if (res.ok) {
					this.issue.status = "merged"
				} else {
					alert((await res.json()).error || res.statusText)
				}
			},
			async closePr() {
				const { user, project, issue } = this.$route.params
				const res = await fetch(`/api/v1/user/${user}/project/${project}/prs/${issue}/close`, {
					method: "POST",
				})
				
				if (res.ok) {
					this.issue.status = "closed"
				} else {
					alert((await res.json()).error || res.statusText)
				}
			},
		}
	}
</script>

<style scoped>
.content {
	white-space: pre-line;
}
</style>
