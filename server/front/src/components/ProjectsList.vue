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
			this.projects = await this.$getData(`/api/v1/user/${this.$route.params.user}/projects`)
		},
	}
</script>

<style scoped>

</style>
