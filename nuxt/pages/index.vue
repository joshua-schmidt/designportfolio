<template>
  <section class="container">
    <div class="mt-5">
      <tile-rows type="project" />
    </div>
  </section>
</template>

<script lang="ts">
import Vue from 'vue'
import TileRows from '~/components/TileRows.vue'
// @ts-ignore
const seo = JSON.parse(process.env.seoconfig)
export default Vue.extend({
  components: {
    TileRows
  },
  data() {
    return {}
  },
  // @ts-ignore
  head() {
    const title = 'Home'
    const description =
      'custom blog and project descriptions for anything in tech and beyond'
    const image = `${seo.url}/icon.jpg`
    return {
      title: title,
      meta: [
        { property: 'og:title', content: title },
        { property: 'og:description', content: description },
        {
          property: 'og:image',
          content: image
        },
        { name: 'twitter:title', content: title },
        {
          name: 'twitter:description',
          content: description
        },
        {
          name: 'twitter:image',
          content: image
        },
        { hid: 'description', name: 'description', content: description }
      ],
      __dangerouslyDisableSanitizers: ['script'],
      script: [
        {
          innerHTML: JSON.stringify({
            '@context': 'http://schema.org',
            '@type': 'WebPage',
            name: title,
            description: description
          }),
          type: 'application/ld+json'
        }
      ]
    }
  },
  mounted() {}
})
</script>

<style lang="scss"></style>
