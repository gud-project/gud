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
						<th scope="row">{{ index + 1 }}</th>
						<td>
							<router-link :to="`/${$route.params.user}/${project}`">{{ project }}</router-link>
						</td>
					</tr>
				</tbody>
		</table>
			<div v-if="isMe">
				<form class="input-group mb-3" v-if="creating" @submit="create">
					<input v-model="info.name" placeholder="Name" pattern="[a-zA-Z0-9_-]+"
						   class="form-control" required />
					<div class="input-group-append">
						<input type="submit" value="Create" class="btn btn-md btn-success" />
					</div>
					<div class="input-group-append">
						<button class="btn btn-md btn-danger" @click="creating = false">Cancel</button>
					</div>
				</form>
				<button class="btn btn-secondary" v-else @click="creating = true">Create Project</button>
			</div>
		</div>
	</div>
</template>

<script>
	export default {
		name: "ProjectsList",
		data() {
			return {
				projects: [],
				creating: false,
				isMe: false,
				info: {
					name: null,
				},
			}
		},
		methods: {
			async create(e) {
				e.preventDefault()

				const res = await fetch("/api/v1/projects/create", {
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify(this.info),
				})
				if (res.ok) {
					this.$router.go(0)
				} else {
					alert((await res.json()).error || res.statusText)
				}
			},
		},
		async created() {
			const user = this.$route.params.user
			this.projects = await this.$getData(`/api/v1/user/${user}/projects`)
			this.isMe = user === await this.$getUser()
		},
	}
</script>

<style scoped>

</style>
