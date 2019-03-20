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

  self.token = undefined;


  self.apiError = function(p_endpoint, p_data) {
    var l_text = "<empty>";
    var l_status = p_data.status + " " + p_data.statusText;
    if (p_data.responseText.length) {
      l_text = p_data.responseText;
    }
    p_app.addError("api error " + p_endpoint);
    p_app.addError("   -> (" + l_status + ") : " + l_text);
  };

  self.init = function(p_callback) {
    self.get("/auth/user_info", function(p_data) {
      if (p_data == undefined) {
        p_data = "";
      }
      self.token = p_data["token"];
      p_callback();
    });
  };

  self.createHeaders = function(p_passToken) {
    var l_headers = {};
    if (p_passToken == undefined)
      p_passToken = true;
    if (true == p_passToken) {
      l_headers["Authorization"] = self.token;
    }
    return l_headers;
  };

  self.get = function(p_endpoint, p_callback, p_passToken) {
    $.ajax({
      url:     p_endpoint,
      type:    "GET",
      headers: self.createHeaders(p_passToken)
    }).
      done(function(p_data) { p_callback(p_data); }).
      fail(function(p_data) {
        self.apiError(p_endpoint, p_data);
      });
  };

  self.postJson = function(p_endpoint, p_data, p_callback, p_passToken) {
    $.ajax({
      url:         p_endpoint,
      data:        JSON.stringify(p_data),
      type:        "POST",
      contentType: "application/json; charset=utf-8",
      dataType:    "json",
      headers:     self.createHeaders(p_passToken)
    }).
      done(function(p_data) { p_callback(p_data); }).
      fail(function(p_data) {
        self.apiError(p_endpoint, p_data);
      });
  };

  self.postMessage = function(p_data, p_callback) {
    Pace.track(function() {
      self.postJson("/v1/message", p_data, p_callback);
    });
  };

  self.postMessageAll = function(p_data, p_callback) {
    Pace.track(function() {
      self.postJson("/v1/message_all", p_data, p_callback);
    });
  };

  self.getOrgs = function(p_callback) {
    Pace.ignore(function() {
      self.get("/v1/orgs", p_callback);
    });
  };

  self.getSpaces = function(p_callback) {
    Pace.ignore(function() {
      self.get("/v1/spaces", p_callback);
    });
  };

  self.getUsers = function(p_callback) {
    Pace.ignore(function() {
      self.get("/v1/users", p_callback);
    });
  };

  self.getServices = function(p_callback) {
    Pace.ignore(function() {
      self.get("/v1/services", p_callback);
    });
  };

  self.getBuildpacks = function(p_callback) {
    Pace.ignore(function() {
      self.get("/v1/buildpacks", p_callback);
    });
  };

  self.getMailCount = function(p_callback) {
    Pace.ignore(function() {
      self.get("/v1/mail/status", p_callback, false);
    });
  };
};


function Message(p_app) {
  var self = this;

  self.md = window.markdownit();

  self.ui = {
    send: $("#msg_send"),
    confirm: $("#msg_confirm"),
    confirm_ok: $("#msg_confirm button.btn-success"),
    msg:  {
      form:    $("#msg_form"),
      subject: $("#msg_subject"),
      content: $("#msg_content")
    },
    preview: {
      content: $("#msg_preview"),
      tab:     $('a[href="#preview"]')
    }
  };

  self.setPreviewContent = function(p_str) {
    var l_content = self.md.render(p_str);
    self.ui.preview.content.html(l_content);
  };

  self.getMsgContent = function() {
      return self.ui.msg.content.val();
  };

  self.onPreviewClick = function() {
    var l_val = self.getMsgContent();
    self.setPreviewContent(l_val);
  };

  self.onSendClick = function() {
    if (true == p_app.targets.targetAll()) {
      self.ui.confirm.modal("show");
    }
    else {
      self.onConfirmClick();
    }
  };

  self.onConfirmClick = function() {
    self.ui.confirm.modal("hide");
    self.ui.msg.form.submit();
  };

  self.onFormSubmit = function(p_event) {
    return self.send();
  };

  self.disableSend = function() {
    self.ui.send.addClass("disabled");
    self.ui.send.attr("disabled", "disabled");
  };

  self.enableSend = function() {
    self.ui.send.removeClass("disabled");
    self.ui.send.removeAttr("disabled");
  };

  self.onMailSent = function(p_data) {
    self.enableSend();
    p_app.addMessage("Mails successfully enqueued.");
    // p_app.addMessage("from: "    + p_data["from"]);
    // p_app.addMessage("subject: " + p_data["subject"]);
    // p_app.addMessage("copy: ");
    // $.each(p_data["copy"], function(c_idx, c_val) {
    //   p_app.addMessage(c_val);
    // });
    // p_app.addMessage("recipients: ");
    // $.each(p_data["recipients"], function(c_idx, c_val) {
    //   p_app.addMessage(c_val);
    // });
  };


  self.saveMessage = function() {
    var l_sub = self.ui.msg.subject.val();
    var l_msg = self.ui.msg.content.val();
    window.localStorage.setItem("subject", l_sub);
    window.localStorage.setItem("message", l_msg);
  };

  self.restoreMessage = function() {
    var l_sub = window.localStorage.getItem("subject");
    var l_msg = window.localStorage.getItem("message");
    if (l_sub != null) {
      self.ui.msg.subject.val(l_sub);
    }
    if (l_msg != null) {
      self.ui.msg.content.val(l_msg);
    }
  };

  self.send = function() {
    if (false == p_app.targets.validate())
      return false;

    var l_data;
    self.disableSend();

    l_data               = p_app.targets.getTargetData();
    l_data["subject"]    = self.ui.msg.subject.val();
    l_data["message"]    = self.ui.msg.content.val();
    l_data["recipients"] = l_data["externals"];
    delete l_data["externals"];

    self.saveMessage();
    if (p_app.targets.targetAll()) {
      delete l_data["orgs"];
      delete l_data["spaces"];
      delete l_data["services"];
      delete l_data["buildpacks"];
      delete l_data["users"];
      p_app.api.postMessageAll(l_data, self.onMailSent);
    } else {
      p_app.api.postMessage(l_data, self.onMailSent);
    }
    return false;
  };

  self.bind = function() {
    self.ui.send.click(self.onSendClick);
    self.ui.preview.tab.click(self.onPreviewClick);
    self.ui.confirm.modal({
      show: false
    });
    self.ui.confirm_ok.click(self.onConfirmClick);
  };

  self.init = function() {
    self.ui.msg.form.validate({
      errorClass: "text-danger",
      invalidHandler: function() {
        p_app.targets.validate();
      },
      submitHandler: self.onFormSubmit
    });

    self.bind();
    self.restoreMessage();
  };

  self.init();
}


function Targets(p_app) {
  var self = this;

  self.ui = {
    all           : $("#tgt-send-all"),
    accordion     : $("#tgt"),
    orgs          : $("#tgt-orgs"),
    spaces        : $("#tgt-spaces"),
    services      : $("#tgt-services"),
    buildpacks    : $("#tgt-buildpacks"),
    users         : $("#tgt-users"),
    externals     : $("#tgt-externals"),
    externals_add : $("#tgt-externals-add"),
    modal         : $("#tgt-mail"),
    modal_add     : $("#tgt-mail button.btn-success"),
    modal_mail    : $("#tgt-mail input"),
    modal_form    : $("#tgt-mail form"),
    error         : $("#tgt-error")
  };


  self.targetAll = function() {
    return self.ui.all.hasClass("active");
  };

  self.validate = function() {
    if (self.targetAll()) {
      self.hideError();
      return true;
    }

    if ($("button[data-id]", self.ui.accordion).length) {
      self.hideError();
      return true;
    }
    self.showError();
    return false;
  };

  self.onAllClick = function(pEvent) {
    self.ui.all.toggleClass("active");
    self.validate();
    pEvent.stopPropagation();
    self.ui.all.blur();
  };

  self.hideError = function() {
    self.ui.error.hide();
  };

  self.showError = function() {
    self.ui.error.show();
  };

  self.getTargetData = function() {
    var l_res   = {
      "orgs"       : [],
      "spaces"     : [],
      "services"   : [],
      "buildpacks" : [],
      "users"      : [],
      "externals"  : []
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
    if (p_type == "externals")  return self.ui.externals;
    return undefined;
  };

  self.removeTarget = function(p_el) {
    p_el.remove();
    self.updateBadges();
    self.validate();
  };

  self.createObject = function(p_type, p_id, p_name) {
    var l_item = $("<li/>", {
      "data-target": p_id,
      "class":       "list-group-item"
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

    l_el.parent().collapse("show");
    self.updateBadges();
    self.validate();
  };

  self.updateBadges = function() {
    $("div.panel", self.ui.accordion).each(function() {
      var l_badge = $(".label", $(this));
      var l_items = $('ul.list-group li', $(this));
      l_badge.html(l_items.length);
    });
  };


  self.onExternalClick = function() {
    self.ui.modal.modal("show");
  };

  self.onModalAddClick = function() {
    if (true == self.ui.modal_form.valid()) {
      self.addTarget("externals", self.ui.modal_mail.val(), self.ui.modal_mail.val());
      self.ui.modal.modal("hide");
      self.ui.modal_mail.val("");
    };
  };

  self.bind = function() {
    self.ui.modal.modal({"show" : false});
    self.ui.modal_add.click(self.onModalAddClick);
    self.ui.externals_add.click(self.onExternalClick);
    self.ui.all.click(self.onAllClick);
  };

  self.init = function() {
    self.hideError();
    self.bind();
    self.ui.modal_form.validate({
      errorClass: "text-danger"
    });
  };

  self.init();
}

function GenericTable(self, p_name, p_app) {
  self.dtable = undefined;

  self.ui    = {
    table : $("#" + p_name + "_table"),
    tab   : $('a[href="#'+ p_name + '"]', $("#app-tabs")),
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
  var self = this;
  GenericTable(self, "orgs", p_app);

  self.initTable = function(p_data) {
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
    self.createTable(p_data, l_cols, self.bind);
  };

  self.onSpaceBtnClick = function() {
    var l_id = $(this).data("id");
    p_app.space.showTab();
    p_app.space.filterOrg(l_id);
  };

  self.bind = function() {
    $('[data-toggle="tooltip"]').tooltip();
    $('[data-tooltip="tooltip"]').tooltip();
    $("button.org_filter", self.ui.table).click(self.onSpaceBtnClick);
    $("button.add_item",   self.ui.table).click(function() {
      p_app.targets.addTarget("orgs", $(this).data("id"), $(this).data("name"));
      $(this).blur();
    });
  };

  self.init = function() {
    p_app.api.getOrgs(self.initTable);
  };

  self.init();
}


function SpaceTable(p_app) {
  var self = this;

  GenericTable(self, "spaces", p_app);

  self.filterOrg = function(p_data) {
    self.ui.filter.val(p_data);
    self.dtable.search(p_data);
    self.dtable.draw();
  };

  self.initTable = function(p_data) {
    if (p_data == undefined) {
      p_data = []
    }

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
    self.createTable(p_data, l_cols, self.bind);
  };

  self.bind = function() {
    $('[data-toggle="tooltip"]').tooltip();
    $("button.add_item",   self.ui.table).click(function() {
      p_app.targets.addTarget("spaces", $(this).data("id"), $(this).data("name"));
      $(this).blur();
    });
  };

  self.init = function() {
    p_app.api.getSpaces(self.initTable);
  };

  self.init();
}


function UserTable(p_app) {
  var self = this;

  GenericTable(self, "users", p_app);

  self.initTable = function(p_data) {
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
    self.createTable(p_data, l_cols, self.bind);
  };

  self.bind = function() {
    $('[data-toggle="tooltip"]').tooltip();
    $("button.add_item",   self.ui.table).click(function() {
      p_app.targets.addTarget("users", $(this).data("id"), $(this).data("name"));
      $(this).blur();
    });
  };

  self.init = function() {
    p_app.api.getUsers(self.initTable);
  };

  self.init();
}


function ServiceTable(p_app) {
  var self = this;

  GenericTable(self, "services", p_app);

  self.initTable = function(p_data) {
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
    self.createTable(p_data, l_cols, self.bind);
  };

  self.bind = function() {
    $('[data-toggle="tooltip"]').tooltip();
    $("button.add_item",   self.ui.table).click(function() {
      p_app.targets.addTarget("services", $(this).data("id"), $(this).data("name"));
      $(this).blur();
    });
  };

  self.init = function() {
    p_app.api.getServices(self.initTable);
  };

  self.init();
}


function BuildpackTable(p_app) {
  var self = this;

  GenericTable(self, "buildpacks", p_app);

  self.initTable = function(p_data) {
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
    self.createTable(p_data, l_cols, self.bind);
  };

  self.bind = function() {
    $('[data-toggle="tooltip"]').tooltip();
    $("button.add_item",   self.ui.table).click(function() {
      p_app.targets.addTarget("buildpacks", $(this).data("id"), $(this).data("name"));
      $(this).blur();
    });
  };

  self.init = function() {
    p_app.api.getBuildpacks(self.initTable);
  };

  self.init();
}


function App() {
  var app = this;

  self.errors = [];
  self.msg    = [];
  self.ui     = {
    error : {
      modal:   $("#app-errors"),
      content: $("#app-errors-content")
    },
    msg : {
      modal:   $("#app-msg"),
      content: $("#app-msg-content")
    },
    mailcount: $("#mailcount")
  };

  self.addError = function(p_error) {
    console.log(p_error);
    self.errors.push($("<div/>").text(p_error).html());
    self.ui.error.content.html(self.errors.join("<br/>"));
    self.ui.error.modal.modal('show');
  };

  self.addMessage = function(p_msg) {
    self.msg.push($("<div/>").text(p_msg).html());
    self.ui.msg.content.html(self.msg.join("<br/>"));
    self.ui.msg.modal.modal('show');
  };

  self.dtAutoFilter = function(p_table) {
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

  self.onUpdateMailCount = function(p_data) {
    var l_count = p_data["outgoing"];
    self.ui.mailcount.html(l_count);
  };

  self.updateMailCount = function() {
    self.api.getMailCount(self.onUpdateMailCount);
  };

  self.onErrorModalHidden = function() {
    self.errors = [];
  };

  self.onMsgModalHidden = function() {
    self.msg = [];
  };

  self.bind = function() {
    self.ui.error.modal.on("hidden.bs.modal", self.onErrorModalHidden);
    self.ui.msg.modal.on("hidden.bs.modal", self.onMsgModalHidden);
  };

  self.initTables = function() {
    self.org       = new OrgTable(self);
    self.space     = new SpaceTable(self);
    self.user      = new UserTable(self);
    self.service   = new ServiceTable(self);
    self.buildpack = new BuildpackTable(self);
    self.org.showTab();
  };

  self.init = function() {
    self.ui.error.modal.modal({ "show" : false });
    self.ui.msg.modal.modal({ "show" : false });
    self.bind();

    self.targets = new Targets(self);
    self.message = new Message(self);
    self.api     = new Api(self);
    self.api.init(self.initTables);

    self.updateMailCount();
    window.setInterval(self.updateMailCount, 60 * 1000);
  };

  self.init();
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
