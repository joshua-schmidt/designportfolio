import Vue from 'vue'
import { library } from '@fortawesome/fontawesome-svg-core'
import {
  faSearch,
  faTimes,
  faAngleDoubleRight,
  faPlusCircle
} from '@fortawesome/free-solid-svg-icons'

import { faStackOverflow } from '@fortawesome/free-brands-svg-icons'

import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

library.add(
  faSearch,
  faStackOverflow,
  faTimes,
  faAngleDoubleRight,
  faPlusCircle
)

Vue.component('font-awesome-icon', FontAwesomeIcon)
