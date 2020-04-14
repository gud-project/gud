<template>
    <div>
        <table class="table">
            <thead class="thead-dark">
                <th id="id" scope="col">#</th>
                <th scope="col">Name</th>
                <th scope="col">Author</th>
                <th scope="col">Created</th>
                <th scope="col">From</th>
                <th scope="col">To</th>
            </thead>
            <tbody>
            <tr v-for="pr in prs">
                <th scope="row">{{ pr.id }}</th>
                <td>
                    <router-link :to="`/${$route.params.user}/${$route.params.project}/pr/${pr.id}`">
                        {{ pr.title }}
                    </router-link>
                </td>
                <td>
                    <router-link :to="`/${pr.id}`">
                        @{{ pr.author }}
                    </router-link>
                </td>
                <td>
					{{ new Date(pr.created).toDateString() }}
				</td>
                <td>
                    {{ pr.from }}
                </td>
                <td>
                    {{ pr.to }}
                </td>
            </tr>
            <tr>
                <td>
                    <router-link class="btn btn-secondary btn-lg" :to="`/${$route.params.user}/${$route.params.project}/pr/new`">
                        add pull request
                    </router-link>
                </td>
            </tr>
            </tbody>
        </table>
    </div>
</template>

<script>
    export default {
        name: "PrsList",
        data() {
            return {
                prs: [],
            }
        },
        async created() {
            const { user, project } = this.$route.params
            this.prs = await this.$getData(`/api/v1/user/${user}/project/${project}/prs`)
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
