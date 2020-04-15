<template>
	<div>
		<table class="table">
			<thead class="thead-dark">
			<th id="id" scope="col">#</th>
			<th scope="col">Name</th>
			<th scope="col">Version</th>
			<th scope="col">Status</th>
			</thead>

			<tbody>
			<tr v-for="job in jobs">
				<th scope="row">
					<router-link :to="`/${$route.params.user}/${$route.params.project}/job/${job.id}`">
						id
					</router-link>
				</th>
				<td>
					{{job.version}}
				</td>
				<td>
					<div :style="`color: ${getStatusColor(job.status)}`">{{ job.status }}</div>
				</td>
			</tr>
			</tbody>
		</table>
	</div>
</template>

<script>
	export default {
		name: "JobsList",
		data() {
			return {
				jobs: [],
			}
		},
		async created() {
			const { user, project } = this.$route.params
			const res = await fetch(`/api/v1/user/${user}/project/${project}/jobs`)
			
			if (res.ok) {
				this.jobs = await res.json()
			} else {
				console.error(res.statusText)
			}
		},
		methods: {
			getStatusColor(status) {
				switch (status) {
					case "pending":
						return "yellow"
					case "success":
						return "green"
					case "failure":
						return "red"
				}
			}
		}
	}
</script>

<style scoped>

</style>
