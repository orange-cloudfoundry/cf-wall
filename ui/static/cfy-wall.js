jQuery.fn.outerHTML = function(s) {
  return (s)
    ? this.before(s).remove()
    : jQuery("<p>").append(this.eq(0).clone()).html();
};

function templateEl(p_el, p_vars, p_callback) {
  var l_content = $(p_el).html();
  for (var c_key in p_vars) {
    var l_regexp  = new RegExp("\\[\\[" + c_key + "\\]\\]", "g");
    l_content = l_content.replace(l_regexp, p_vars[c_key]);
  }

  var l_el = $(l_content);
  if (undefined != p_callback)
    p_callback(l_el);
  return l_el;
}

function template(p_el, p_vars, p_callback, p_str) {
  var l_el = templateEl(p_el, p_vars, p_callback);
  return $(l_el).outerHTML();
};

function Api(p_app) {
  var self = this;

  self.get = function(p_endpoint, p_callback) {
    $.ajax({
      url : p_endpoint,
      type : "GET"
    }).
      done(function(p_data) { p_callback(p_data) }).
      fail(function()       {
        p_app.showError("error on api endpoint : " + p_endpoint);
        p_callback([]);
      });
  };

  self.postJson = function(p_endpoint, p_data, p_callback) {
    $.ajax({
      url : p_endpoint,
      data: JSON.stringify(p_data),
      type : "POST",
      contentType:"application/json; charset=utf-8",
      dataType:"json"
    }).
      done(function(p_data) { p_callback(p_data) }).
      fail(function()       {
        p_app.showError("error on api endpoint : " + p_endpoint);
      });
  };

  self.postMessage = function(p_data) {
    self.postJson("/v1/message", p_data, function(p_out) {
      p_app.showMessage(JSON.stringify(p_out));
    });
  };

  self.getOrgs = function(p_callback) {
    return self.get("/v1/orgs", p_callback);
  };

  self.getSpaces = function(p_callback) {
    return self.get("/v1/spaces", p_callback);
  };

  self.getUsers = function(p_callback) {
    return self.get("/v1/users", p_callback);
  };

  self.getServices = function(p_callback) {
    return self.get("/v1/services", p_callback);
  };

  self.getBuildpacks = function(p_callback) {
    return self.get("/v1/buildpacks", p_callback);
  };
};


function Message(p_app) {
  var self = this;

  self.ui = {
    form    : $("#form"),
    send    : $("#send"),
    subject : $("#message_subject"),
    message : $("#message_message")
  };

  self.onSendClick = function() {
    self.ui.form.submit();
  };

  self.bind = function() {
    self.ui.send.click(self.onSendClick);
  };

  self.onFormSubmit = function(p_event) {
    if (false == p_event.isDefaultPrevented()) {
      self.send();
    }
    return false;
  };

  self.send = function() {
    var l_data = p_app.targets.getTargetData();
    l_data["subject"] = self.ui.subject.val();
    l_data["message"] = self.ui.message.val();
    p_app.api.postMessage(l_data);
  };

  self.init = function() {
    self.ui.form.validator();
    self.ui.form.on("submit", self.onFormSubmit);
    self.bind();
  };

  self.init();
}


function Targets(p_app) {
  var self = this;

  self.ui = {
    accordion  : $("#targets"),
    orgs       : $("#target_orgs"),
    spaces     : $("#target_spaces"),
    services   : $("#target_services"),
    buildpacks : $("#target_buildpacks"),
    users      : $("#target_users")
  };

  self.getTargetData = function() {
    var l_res   = {
      "orgs"       : [],
      "spaces"     : [],
      "services"   : [],
      "buildpacks" : [],
      "users"      : []
    };
    $("button[data-id]", self.ui.accordion).each(function() {
      var l_type = $(this).data("type");
      var l_id   = $(this).data("id");
      l_res[l_type].push(l_id);
    });
    return l_res;
  };

  self.getTypeTarget = function(p_type) {
    if (p_type == "orgs")       return self.ui.orgs;
    if (p_type == "spaces")     return self.ui.spaces;
    if (p_type == "services")   return self.ui.services;
    if (p_type == "buildpacks") return self.ui.buildpacks;
    if (p_type == "users")      return self.ui.users;
    return undefined;
  };

  self.removeTarget = function(p_el) {
    p_el.remove();
    self.updateBadges();
  };

  self.createObject = function(p_type, p_id, p_name) {
    var l_item = $("<li/>", {
      "data-target" : p_id
    });
    var l_el = templateEl($("#tpl-target"), {
      name : p_name,
      id : p_id,
      type : p_type
    });
    l_item.append(l_el);
    $("button", l_el).click(function() {
      self.removeTarget(l_item);
    });
    return l_item;
  };


  self.addTarget = function(p_type, p_id, p_name) {
    var l_el   = self.getTypeTarget(p_type);
    var l_item = $('li[data-target="' + p_id + '"]', l_el);
    if (l_item.length == 0) {
      l_item = self.createObject(p_type, p_id, p_name);
      l_el.append(l_item);
      $('[data-toggle="tooltip"]', l_el).tooltip();
    }
    l_el.collapse("show");
    self.updateBadges();
  };

  self.updateBadges = function() {
    $("li.panel", self.ui.accordion).each(function() {
      var l_badge = $(".label", $(this));
      var l_items = $('ul li', $(this));
      l_badge.html(l_items.length);
    });
  };
}

function GenericTable(self, p_name, p_app) {
  self.dtable = undefined;

  self.ui    = {
    table : $("#" + p_name + "_table"),
    tab   : $('a[href="#'+ p_name + '"]', p_app.ui.tabs),
    tab_content : $("#" + p_name),
    filter : undefined
  };

  self.showTab = function() {
    self.ui.tab.tab("show");
  };

  self.addActionColumn = function(p_data) {
    var l_data = [];
    $.each(p_data, function(c_idx, c_val) {
      c_val["actions"] = "";
      if (c_val["name"] != "") {
        l_data.push(c_val);
      }
    });
    return l_data;
  };

  self.createTable = function(p_data, p_columns, p_drawCallback) {
    self.dtable = self.ui.table.DataTable({
      lengthMenu: [ 5, 10, 20, 40, 60, 100 ],
      paging:   true,
      ordering: true,
      info:     false,
      searching : true,
      lengthChange: true,
      data: self.addActionColumn(p_data),
      columns: p_columns
    });
    p_app.dtAutoFilter(self.dtable);
    self.ui.filter = $("#"+ p_name + "_table_filter input");
    self.dtable.on("draw", p_drawCallback);
    p_drawCallback();
  };
}


function OrgTable(p_app) {
  var org = this;
  GenericTable(org, "orgs", p_app);

  org.initTable = function(p_data) {
    var l_cols = [
        {
          "data"      : "name",
          "className" : "text-center"
        },
        {
          "data"      : "guid",
          "className" : "text-center"
        },
        {
          "data" :  "actions",
          "render" : function(p_data, p_type, p_row, p_meta) {
            return template($("#tpl-org-btn"), p_row);
          },
          "className" : "text-center"
        }
    ];
    org.createTable(p_data, l_cols, org.bind);
  };

  org.onSpaceBtnClick = function() {
    var l_id = $(this).data("id");
    p_app.space.showTab();
    p_app.space.filterOrg(l_id);
  };

  org.bind = function() {
    $('[data-toggle="tooltip"]').tooltip();
    $("button.org_filter", org.ui.table).click(org.onSpaceBtnClick);
    $("button.add_item",   org.ui.table).click(function() {
      p_app.targets.addTarget("orgs", $(this).data("id"), $(this).data("name"));
    });
  };

  org.init = function() {
    p_app.api.getOrgs(org.initTable);
  };

  org.init();
}



function SpaceTable(p_app) {
  var space = this;

  GenericTable(space, "spaces", p_app);

  space.filterOrg = function(p_data) {
    space.ui.filter.val(p_data);
    space.dtable.search(p_data);
    space.dtable.draw();
  };

  space.initTable = function(p_data) {
    var l_cols = [
        {
          "data"      : "name",
          "className" : "text-center"
        },
        {
          "data"      : "guid",
          "className" : "text-center",
          "visible"   : true
        },
        {
          "data"      : "org_guid",
          "className" : "text-center",
          "visible"   : true
        },
        {
          "data" :  "actions",
          "render" : function(p_data, p_type, p_row, p_meta) {
            return template($("#tpl-space-btn"), p_row);
          },
          "className" : "text-center"
        }
    ];
    space.createTable(p_data, l_cols, space.bind);
  };

  space.bind = function() {
    $('[data-toggle="tooltip"]').tooltip();
    $("button.add_item",   space.ui.table).click(function() {
      p_app.targets.addTarget("spaces", $(this).data("id"), $(this).data("name"));
    });
  };

  space.init = function() {
    p_app.api.getSpaces(space.initTable);
  };

  space.init();
}

function UserTable(p_app) {
  var user = this;

  GenericTable(user, "users", p_app);

  user.initTable = function(p_data) {
    var l_cols = [
        {
          "data"      : "name",
          "className" : "text-center"
        },
        {
          "data"      : "guid",
          "className" : "text-center",
          "visible"   : true
        },
        {
          "data" :  "actions",
          "render" : function(p_data, p_type, p_row, p_meta) {
            return template($("#tpl-user-btn"), p_row);
          },
          "className" : "text-center"
        }
    ];
    user.createTable(p_data, l_cols, user.bind);
  };

  user.bind = function() {
    $('[data-toggle="tooltip"]').tooltip();
    $("button.add_item",   user.ui.table).click(function() {
      p_app.targets.addTarget("users", $(this).data("id"), $(this).data("name"));
    });
  };

  user.init = function() {
    p_app.api.getUsers(user.initTable);
  };

  user.init();
}

function ServiceTable(p_app) {
  var service = this;

  GenericTable(service, "services", p_app);

  service.initTable = function(p_data) {
    var l_cols = [
        {
          "data"      : "name",
          "className" : "text-center"
        },
        {
          "data"      : "guid",
          "className" : "text-center",
          "visible"   : true
        },
        {
          "data" :  "actions",
          "render" : function(p_data, p_type, p_row, p_meta) {
            return template($("#tpl-service-btn"), p_row);
          },
          "className" : "text-center"
        }
    ];
    service.createTable(p_data, l_cols, service.bind);
  };

  service.bind = function() {
    $('[data-toggle="tooltip"]').tooltip();
    $("button.add_item",   service.ui.table).click(function() {
      p_app.targets.addTarget("services", $(this).data("id"), $(this).data("name"));
    });
  };

  service.init = function() {
    p_app.api.getServices(service.initTable);
  };

  service.init();
}

function BuildpackTable(p_app) {
  var buildpack = this;

  GenericTable(buildpack, "buildpacks", p_app);

  buildpack.initTable = function(p_data) {
    var l_cols = [
        {
          "data"      : "name",
          "className" : "text-center"
        },
        {
          "data"      : "guid",
          "className" : "text-center",
          "visible"   : true
        },
        {
          "data" :  "actions",
          "render" : function(p_data, p_type, p_row, p_meta) {
            return template($("#tpl-buildpack-btn"), p_row);
          },
          "className" : "text-center"
        }
    ];
    buildpack.createTable(p_data, l_cols, buildpack.bind);
  };

  buildpack.bind = function() {
    $('[data-toggle="tooltip"]').tooltip();
    $("button.add_item",   buildpack.ui.table).click(function() {
      p_app.targets.addTarget("buildpacks", $(this).data("id"), $(this).data("name"));
    });
  };

  buildpack.init = function() {
    p_app.api.getBuildpacks(buildpack.initTable);
  };

  buildpack.init();
}

function App() {
  var app = this;

  app.errors = [];

  app.showMessage = function(p_message) {
    alert(p_message);
    // todo
  };

  app.showError = function(p_error) {
    app.errors.push(p_error);
    // todo
  };

  app.ui = {
    tabs : $("#object-tabs")
  };

  app.dtAutoFilter = function(p_table) {
    p_table.columns().every(function() {
      var l_col = this;
      $('input', this.footer()).on('keyup change', function() {
        if (l_col.search() !== this.value) {
          l_col
            .search(this.value)
            .draw();
        }
      });
    });
  };


  app.init = function() {
    app.targets   = new Targets(app);
    app.api       = new Api(app);
    app.org       = new OrgTable(app);
    app.space     = new SpaceTable(app);
    app.user      = new UserTable(app);
    app.service   = new ServiceTable(app);
    app.buildpack = new BuildpackTable(app);
    app.msg       = new Message(app);

    app.org.showTab();
  };

  app.init();
}
