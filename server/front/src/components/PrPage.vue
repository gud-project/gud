<template>
	<div class="container">
		<div class="jumbotron">
			<h1>!{{ $route.params.pr }} : {{ pr.title }}</h1>
			<label>
				<router-link :to="pr.author">@{{ pr.author }}</router-link>
			</label>
			<br />
			<p class="larger-text"> {{ pr.from }} &rarr; {{ pr.to }}</p>
			<p class="content">{{ pr.content }}</p>
			<p class="larger-text">Current Status: <b>{{ pr.status }}</b></p>
			<div class="btn-group">
				<button @click="mergePr" :disabled="pr.status !== 'open'" class="btn btn-success">Merge</button>
				<button @click="closePr" :disabled="pr.status !== 'open'" class="btn btn-danger">Close</button>
			</div>
			<br /><br />
			<p>{{ new Date(pr.created).toDateString() }}</p>
		</div>
	</div>
</template>

<script>
	export default {
		name: "PrPage",
		data() {
			return {
				pr: {
					title: null,
					author: null,
					content: null,
					status: null,
					from: null,
					to: null,
					created: null,
				},
			}
		},
		async created() {
			const { user, project, pr } = this.$route.params
			this.pr = await this.$getData(
				`/api/v1/user/${user}/project/${project}/prs/${pr}`)
			this.status = this.pr.status
		},
		methods: {
			async mergePr() {
				const { user, project, pr } = this.$route.params
				const res = await fetch(`/api/v1/user/${user}/project/${project}/prs/${pr}/merge`, {
					method: "POST",
				})
				
				if (res.ok) {
					this.pr.status = "merged"
				} else {
					alert((await res.json()).error || res.statusText)
				}
			},
			async closePr() {
				const { user, project, pr } = this.$route.params
				const res = await fetch(`/api/v1/user/${user}/project/${project}/prs/${pr}/close`, {
					method: "POST",
				})
				
				if (res.ok) {
					this.pr.status = "closed"
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
.larger-text {
	font-size: larger;
}
</style>
