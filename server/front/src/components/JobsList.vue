<template>
	<div>
		<div v-for="job in jobs">
			<router-link :to="`/${$route.params.user}/${$route.params.project}/job/${job.id}`">
				#{{ job.id }}
			</router-link>
			<div :style="`color: ${getStatusColor(job.status)}`">{{ job.status }}</div>
		</div>
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
