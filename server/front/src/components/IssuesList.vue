<template>
	<div>
		<table class="table">
			<thead class="thead-dark">
				<th id="id" scope="col">#</th>
				<th scope="col">Name</th>
				<th scope="col">Author</th>
				<th scope="col">State</th>
			</thead>
			<tbody>
			<tr v-for="issue in issues">
				<th scope="row">{{ issue.id }}</th>
				<td>
					<router-link :to="`/${$route.params.user}/${$route.params.project}/issue/${issue.id}`">
						{{ issue.title }}
					</router-link>
				</td>
				<td>
					<router-link :to="`/${issue.id}`">
						@{{ issue.author }}
					</router-link>
				</td>
				<td>
					{{ issue.state }}
				</td>
			</tr>
			<tr>
				<td>
					<router-link class="btn btn-secondary btn-lg" :to="`/${$route.params.user}/${$route.params.project}/issue/new`">
						add issue
					</router-link>
				</td>
			</tr>
			</tbody>
		</table>
	</div>
</template>

<script>
	export default {
		name: "IssuesList",
		data() {
			return {
				issues: [],
			}
		},
		async created() {
			const { user, project } = this.$route.params
			this.issues = await this.$getData(`/api/v1/user/${user}/project/${project}/issues`)
		}
	}
</script>

<style scoped>
	th{
		width:250px;
	}

	#id{
		width:50px
	}
</style>
