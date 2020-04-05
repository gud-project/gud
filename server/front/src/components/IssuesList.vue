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
					<router-link :to="`/${$route.params.user}/${$route.params.project}/${category}/${issue.id}`">
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
					<router-link class="btn btn-secondary btn-lg" :to="`/${$route.params.user}/${$route.params.project}/${category}/new`">
						add {{ category }}
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
		props: {
			category: {
				type: String,
				default: 'issue',
			},
		},
		data() {
			return {
				issues: [],
			}
		},
		async created() {
			const { user, project } = this.$route.params
			const res = await fetch(`/api/v1/user/${user}/project/${project}/${this.category}s`)
			
			if (res.ok) {
				this.issues = await res.json()
			} else {
				console.error(res.statusText)
			}
		},
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
