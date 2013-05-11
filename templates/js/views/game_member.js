window.GameMemberView = Backbone.View.extend({

  tagName: 'select',

  template: _.template($('#game_member_underscore').html()),

	initialize: function() {
	  _.bindAll(this, 'render');
		this.model.bind('change', this.render);
	},

  render: function() {
    this.$el.html(this.template({
		  model: this.model,
		}));
		return this;
	},

});
