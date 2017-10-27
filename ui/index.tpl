<html>
  <head>
    <title> CloudFoundry Wall Ui </title>

    <script src="/ui/static/bower_components/jquery/dist/jquery.min.js"></script>
    <script src="/ui/static/bower_components/bootstrap/dist/js/bootstrap.min.js"></script>
    <script src="/ui/static/bower_components/datatables.net/js/jquery.dataTables.min.js"></script>
    <script src="/ui/static/bower_components/datatables.net-bs/js/dataTables.bootstrap.min.js"></script>
    <script src="/ui/static/bower_components/bootstrap-validator/dist/validator.js"></script>
    <script src="/ui/static/cfy-wall.js"></script>
    <link rel="stylesheet" href="/ui/static/bower_components/bootstrap/dist/css/bootstrap.min.css"/>
    <link rel="stylesheet" href="/ui/static/bower_components/datatables.net-bs/css/dataTables.bootstrap.min.css"/>
    <link rel="stylesheet" href="/ui/static/bower_components/font-awesome/css/font-awesome.min.css" type="text/css" media="all" />
    <link rel="stylesheet" href="/ui/static/style.css" type="text/css" media="all" />
    <script>
     $(document).ready(function(){
       var g_app = new App();
     });
    </script>

    <style>
     #targets li {
       padding-top: 5px;
       list-style: none;
     }
     #targets li.panel {
       margin-bottom: 2px;
     }
    </style>
  </head>
  <body>
    {{ template "header.tpl" }}
    <div class="container-fluid">
      <div class="row">
          <ul id="object-tabs" class="nav nav-tabs" role="tablist">
            <li role="presentation"> <a href="#orgs"       aria-controls="orgs"       role="tab" data-toggle="tab">Orgs</a></li>
            <li role="presentation"> <a href="#spaces"     aria-controls="spaces"     role="tab" data-toggle="tab">Spaces</a></li>
            <li role="presentation"> <a href="#services"   aria-controls="services"   role="tab" data-toggle="tab">Services</a></li>
            <li role="presentation"> <a href="#buildpacks" aria-controls="buildpacks" role="tab" data-toggle="tab">Build Packs</a></li>
            <li role="presentation"> <a href="#users"      aria-controls="users"      role="tab" data-toggle="tab">Users</a></li>
          </ul>
          <div class="tab-content objects">
            {{ template "table.tpl" mkDict "Id" "orgs"       "Cols" (mkSlice "Name" "Guid")         }}
            {{ template "table.tpl" mkDict "Id" "spaces"     "Cols" (mkSlice "Name" "Guid" "Org")   }}
            {{ template "table.tpl" mkDict "Id" "services"   "Cols" (mkSlice "Name" "Guid") }}
            {{ template "table.tpl" mkDict "Id" "buildpacks" "Cols" (mkSlice "Name" "Guid") }}
            {{ template "table.tpl" mkDict "Id" "users"      "Cols" (mkSlice "Name" "Guid")         }}
          </div>
      </div>

      <div class="row">
        <div class="col-md-3">
          <div class="row text-center"><span class="h4">Targets</span></div>
          <div class="row">
          <ul class="nav nav-stacked" id="targets">
            <li class="panel">
              <a data-toggle="collapse" data-parent="#targets" href="#target_orgs">Orgs<span class="pull-right label label-primary">0</span></a>
              <ul id="target_orgs" class="collapse"></ul>
            </li>
            <li class="panel">
              <a data-toggle="collapse" data-parent="#targets" href="#target_spaces">Spaces<span class="pull-right label label-primary">0</span></a>
              <ul id="target_spaces" class="target_section collapse"></ul>
            </li>
            <li class="panel">
              <a data-toggle="collapse" data-parent="#targets" href="#target_services">Services<span class="pull-right label label-primary">0</span></a>
              <ul id="target_services" class="collapse"></ul>
            </li>
            <li class="panel">
              <a data-toggle="collapse" data-parent="#targets" href="#target_buildpacks">Build Packs<span class="pull-right label label-primary">0</span></a>
              <ul id="target_buildpacks" class="collapse"></ul>
            </li>
            <li class="panel">
              <a data-toggle="collapse" data-parent="#targets" href="#target_users">Users<span class="pull-right label label-primary">0</span></a>
              <ul id="target_users" class="collapse"></ul>
            </li>
          </ul>
          </div>
        </div>
        <div class="col-md-9">
          <ul class="nav nav-tabs" role="tablist">
            <li role="presentation" class="active"><a href="#message" aria-controls="message" role="tab" data-toggle="tab">Message</a></li>
            <li role="presentation">               <a href="#preview" aria-controls="preview" role="tab" data-toggle="tab">Preview</a></li>
          </ul>
          <div class="tab-content">
            <div role="tabpanel" class="tab-pane active" id="message">
              <br/>
              <form id="form" data-toggle="validator" role="form">
                <div class="form-group">
                  <input type="text" class="form-control" id="message_subject" placeholder="Subject..." required>
                  <div class="help-block with-errors"></div>
                </div>
                <div class="form-group has-feedback">
                  <textarea style="min-height:200px;" id="message_message" class="form-control" placeholder="Markdown message..." required></textarea>
                  <div class="help-block with-errors"></div>
                </div>
              </form>
            </div>
            <div role="tabpanel" class="tab-pane" id="preview">preview here</div>
          </div>

          <div class="row text-center">
            <br/>
            <button id="send" class="btn btn-success">
              <span class="glyphicon glyphicon-envelope"></span>Send
            </button>
            <br/>
          </div>
        </div>
      </div>


    </div>


    <div class="hidden" id="tpl-org-btn">
      <div class="btn-group">
        <button data-toggle="tooltip" data-placement="top" title="Add organization" class='btn btn-success btn-xs glyphicon glyphicon-check add_item' data-id='[[guid]]' data-name='[[name]]'></button>
        <button data-toggle="tooltip" data-placement="top" title="Search spaces"  class='btn btn-primary btn-xs glyphicon glyphicon-arrow-right org_filter' data-id='[[guid]]'></button>
      </div>
    </div>


    <div class="hidden" id="tpl-space-btn">
      <div class="btn-group">
        <button data-toggle="tooltip" data-placement="top" title="Add space" class='btn btn-success btn-xs glyphicon glyphicon-check add_item' data-id='[[guid]]' data-name='[[name]]'></button>
      </div>
    </div>

    <div class="hidden" id="tpl-user-btn">
      <div class="btn-group">
        <button data-toggle="tooltip" data-placement="top" title="Add user" class='btn btn-success btn-xs glyphicon glyphicon-check add_item' data-id='[[guid]]' data-name='[[name]]'></button>
      </div>
    </div>

    <div class="hidden" id="tpl-service-btn">
      <div class="btn-group">
        <button data-toggle="tooltip" data-placement="top" title="Add service" class='btn btn-success btn-xs glyphicon glyphicon-check add_item' data-id='[[guid]]' data-name='[[name]]'></button>
      </div>
    </div>

    <div class="hidden" id="tpl-buildpack-btn">
      <div class="btn-group">
        <button data-toggle="tooltip" data-placement="top" title="Add buildpack" class='btn btn-success btn-xs glyphicon glyphicon-check add_item' data-id='[[guid]]' data-name='[[name]]'></button>
      </div>
    </div>

    <div class="hidden" id="tpl-target">
      <span>
        <button data-toggle="tooltip" data-placement="top" title="Remove target" data-id="[[id]]" data-type="[[type]]" class="btn btn-danger btn-xs glyphicon glyphicon-remove"></button>
        [[name]]
      </span>
    </div>

  </body>
</html>
