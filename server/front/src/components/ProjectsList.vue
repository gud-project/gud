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
						<td>
							<button class="btn btn-secondary" id="create-project-button" onclick="createProject(true)">Create Project</button>
							<form @submit="create" class="input-group mb-3" id="create-project-form" hidden>
								<input id="new-project-name" v-model="info.name" placeholder="Name" pattern="[a-zA-Z0-9_-]+" class="form-control" required />
								<div class="input-group-append">
									<input type="submit" value="Save" class="btn btn-md btn-success"/>
								</div>
								<div class="input-group-append">
									<button class="btn btn-md btn-danger" onclick="createProject(false)">Cancel</button>
								</div>
							</form>
						</td>
					</tr>
				</tbody>
		</table>
		</div>
	</div>
</template>

<script>
	function createProject(create) {
		const button = document.getElementById("create-project-button");
		const form = document.getElementById("create-project-form");
		if (create) {
			button.style.visibility = "hidden";
			form.style.visibility = "visible";
		} else {
			button.style.visibility = "visible";
			form.style.visibility = "hidden";
		}
	}

	export default {
		name: "ProjectsList",
		data() {
			return {
				projects: [],
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
			this.projects = await this.$getData(`/api/v1/user/${this.$route.params.user}/projects`)
		},
	}
</script>

<style scoped>
#new-project-name {
	width:1px
}
</style>
