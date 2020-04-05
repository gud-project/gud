<template>
	<div class="container">
		<div class="jumbotron">

		<h2 class="text-left">Your projects:</h2>
		<table class="table">

			<thead class="thead-dark">
			<th scope="col">#</th>
			<th scope="col">Name</th>
			</thead>
				<tbody>
					<tr v-for="(project, index) in projects" v-bind:key="project">
						<th scope="row">{{ index+1 }}</th>
						<td>
							<router-link :to="`/${$route.params.user}/${project}`">{{ project }}</router-link>
						</td>
					</tr>
					<tr>
						<router-link class="btn btn-secondary btn-lg" :to="`/${$route.params.user}/${$route.params.project}/${category}/new`">
							new project
						</router-link>
					</tr>
				</tbody>
		</table>
		</div>
	</div>
</template>

<script>
	export default {
		name: "ProjectsList",
		data() {
			return {
				projects: [],
			}
		},
		async created() {
			const res = await fetch(`/api/v1/user/${this.$route.params.user}/projects`)
			if (res.ok) {
				this.projects = await res.json()
			} else {
				console.error(res.statusText)
			}
		},
	}
</script>

<style scoped>

</style>
