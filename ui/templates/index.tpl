<html lang="en">
  <head>
    <title> CloudFoundry Wall Ui </title>
    <script src="/ui/static/bower_components/jquery/dist/jquery.min.js"></script>
    <script src="/ui/static/bower_components/bootstrap/dist/js/bootstrap.min.js"></script>
    <script src="/ui/static/bower_components/datatables.net/js/jquery.dataTables.min.js"></script>
    <script src="/ui/static/bower_components/datatables.net-bs/js/dataTables.bootstrap.min.js"></script>
    <script src="/ui/static/bower_components/jquery-validation/dist/jquery.validate.min.js"></script>
    <script src="/ui/static/bower_components/markdown-it/dist/markdown-it.min.js"></script>
    <script src="/ui/static/bower_components/jquery.cookie/jquery.cookie.js"></script>
    <script src="/ui/static/cf-wall.js"></script>
    <link rel="shortcut icon" href="/ui/static/favicon.ico" />
    <link rel="stylesheet"    href="/ui/static/bower_components/bootstrap/dist/css/bootstrap.min.css"/>
    <link rel="stylesheet"    href="/ui/static/bower_components/datatables.net-bs/css/dataTables.bootstrap.min.css"/>
    <link rel="stylesheet"    href="/ui/static/bower_components/font-awesome/css/font-awesome.min.css" type="text/css" media="all" />
    <link rel="stylesheet"    href="/ui/static/bower_components/PACE/themes/blue/pace-theme-corner-indicator.css" type="text/css" media="all" />
    <link rel="stylesheet"    href="/ui/static/style.css" type="text/css" media="all" />
    <script>
     window.paceOptions = {
       document: false, // disabled
       eventLag: false, // disabled
       restartOnPushState: false,
       restartOnRequestAfter: false,
       startOnPageLoad: false
     }

     $(document).ready(function(){
       var g_app = new App();
     });
    </script>

    <script src="/ui/static/bower_components/PACE/pace.min.js"></script>

    <style>
     .modal-content
     {
       border-bottom-left-radius: 6px;
       border-bottom-right-radius: 6px;
       -webkit-border-bottom-left-radius: 6px;
       -webkit-border-bottom-right-radius: 6px;
       -moz-border-radius-bottomleft: 6px;
       -moz-border-radius-bottomright: 6px;
     }

     .modal-header
     {
       border-top-left-radius: 6px;
       border-top-right-radius: 6px;
       -webkit-border-top-left-radius: 6px;
       -webkit-border-top-right-radius: 6px;
       -moz-border-radius-topleft: 6px;
       -moz-border-radius-topright: 6px;
     }

     div.dataTables_filter {
       width: 100%;
     }
     div.dataTables_filter label {
       width:100%;
     }
     div.dataTables_filter label input {
       width:80% !important;
     }
     #app-msg-content, #app-errors-content {
       max-height: calc(100vh - 110px);
       overflow-y: scroll;
     }
    </style>
  </head>
  <body>
    {{ template "header.tpl" }}
    <div class="container-fluid">
      <div class="row">
        <ul id="app-tabs" class="nav nav-tabs" role="tablist">
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
        <!-- Targets  -->
        <div class="col-md-3">
          <div class="row" >
            <div class="col-xs-4 col-xs-offset-2" style="text-align:center;">
              <button id="tgt-send-all" class="btn btn-lg btn-default fa fa-users" data-tooltip="tooltip" data-placement="top" title="Send to everyone" data-toggle="button" aria-pressed="false" autocomplete="off"/>
            </div>
            <div class="col-xs-4" style="text-align:center;">
              <button id="tgt-externals-add" class="btn btn-lg  btn-success fa fa-plus"  data-toggle="tooltip" data-placement="top" title="Add external target"/>
            </div>
          </div>
          <br/>
          <div class="row">
            <div class="panel-group" id="tgt" role="tablist" aria-multiselectable="true">
              {{ template "accordion.tpl" mkDict "Name" "orgs"       "Title" "Organizations" }}
              {{ template "accordion.tpl" mkDict "Name" "spaces"     "Title" "Spaces"        }}
              {{ template "accordion.tpl" mkDict "Name" "services"   "Title" "Services"      }}
              {{ template "accordion.tpl" mkDict "Name" "buildpacks" "Title" "Build Packs"   }}
              {{ template "accordion.tpl" mkDict "Name" "users"      "Title" "Users"         }}
              {{ template "accordion.tpl" mkDict "Name" "externals"  "Title" "Externals"     }}
            </div>
            <div class="form-group has-error has-danger text-center">
              <label id="tgt-error" class="text-danger" for="msg_subject">You must add at least one target.</label>
            </div>
          </div>
        </div>

        <!-- Message -->
        <div class="col-md-9">
          <ul class="nav nav-tabs" role="tablist">
            <li role="presentation" class="active"><a href="#message" aria-controls="message" role="tab" data-toggle="tab">Message</a></li>
            <li role="presentation">               <a href="#preview" aria-controls="preview" role="tab" data-toggle="tab">Preview</a></li>
          </ul>
          <div class="tab-content">
            <div role="tabpanel" class="tab-pane active" id="message">
              <br/>
              <form id="msg_form" role="form">
                <div class="form-group">
                  <input name="subject" type="text" class="required form-control" id="msg_subject" placeholder="Subject...">
                </div>
                <div class="form-group has-feedback">
                  <textarea style="min-height:200px;" id="msg_content" name="message" class="required form-control" placeholder="Markdown message..."></textarea>
                </div>
              </form>
            </div>
            <div role="tabpanel" class="tab-pane" id="preview">
              <br/>
              <div id="msg_preview">
              </div>
            </div>
          </div>

          <div class="row">
            <br/>
            <button id="msg_send" class="btn btn-success btn-large col-xs-2 col-xs-offset-5">
              <span class="glyphicon glyphicon-envelope pull-left"></span>
              Send
            </button>
          </div>
        </div>
      </div>

    </div>

    <div id="msg_confirm" class="modal fade" tabindex="-1" role="dialog">
      <div class="modal-dialog" role="document">
        <div class="modal-content">
          <div class="modal-body text-center">
            This will send an email to all users,<br/>
            Are you sure ?
          </div>
          <div class="modal-footer">
            <div class="text-center">
              <div class="btn-group">
                <button class="btn btn-danger" data-dismiss="modal">Cancel</button>
                <button class="btn btn-success">Confirm</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div id="tgt-mail" class="modal fade" tabindex="-1" role="dialog">
      <div class="modal-dialog" role="document">
        <div class="modal-content">
          <div class="modal-body">
            <form class="form-horizontal">
              <div class="form-group">
                <label class="col-xs-2 control-label" for="email">Email</label>
                <div class="col-xs-10">
                  <input type="email" class="required email form-control" id="email" placeholder="name@domain.com">
                </div>
              </div>
            </form>
          </div>
          <div class="modal-footer">
            <div class="text-center">
              <div class="btn-group">
                <button class="btn btn-danger" data-dismiss="modal">Cancel</button>
                <button class="btn btn-success">Confirm</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>


    <div id='app-errors' class="modal fade" tabindex="-1" role="dialog">
      <div class="modal-dialog" role="document">
        <div class="modal-content">
          <div class="modal-header alert-danger">
            <button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>
            <h4 class="modal-title">API Errors</h4>
          </div>
          <div class="modal-body" id='app-errors-content'>
          </div>
        </div>
      </div>
    </div>

    <div id='app-msg' class="modal fade" tabindex="-1" role="dialog">
      <div class="modal-dialog" role="document">
        <div class="modal-content">
          <div class="modal-header alert-success">
            <button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>
            <h4 class="modal-title">Messages</h4>
          </div>
          <div class="modal-body" id='app-msg-content'>
          </div>
        </div>
      </div>
    </div>


    <!-- Templates -->
    <div class="hidden" id="tpl-org-btn">
      <div class="btn-group">
        <button data-toggle="tooltip" data-placement="right" title="Add organization" class='btn btn-success btn-xs glyphicon glyphicon-check add_item' data-id='[[guid]]' data-name='[[name]]'></button>
        <button data-toggle="tooltip" data-placement="right" title="Search spaces"  class='btn btn-primary btn-xs glyphicon glyphicon-arrow-right org_filter' data-id='[[guid]]'></button>
      </div>
    </div>
    <div class="hidden" id="tpl-space-btn">
      <div class="btn-group">
        <button data-toggle="tooltip" data-placement="right" title="Add space" class='btn btn-success btn-xs glyphicon glyphicon-check add_item' data-id='[[guid]]' data-name='[[name]]'></button>
      </div>
    </div>
    <div class="hidden" id="tpl-user-btn">
      <div class="btn-group">
        <button data-toggle="tooltip" data-placement="right" title="Add user" class='btn btn-success btn-xs glyphicon glyphicon-check add_item' data-id='[[guid]]' data-name='[[name]]'></button>
      </div>
    </div>
    <div class="hidden" id="tpl-service-btn">
      <div class="btn-group">
        <button data-toggle="tooltip" data-placement="right" title="Add service" class='btn btn-success btn-xs glyphicon glyphicon-check add_item' data-id='[[guid]]' data-name='[[name]]'></button>
      </div>
    </div>
    <div class="hidden" id="tpl-buildpack-btn">
      <div class="btn-group">
        <button data-toggle="tooltip" data-placement="right" title="Add buildpack" class='btn btn-success btn-xs glyphicon glyphicon-check add_item' data-id='[[guid]]' data-name='[[name]]'></button>
      </div>
    </div>
    <div class="hidden" id="tpl-target">
      <span>
        <button data-toggle="tooltip" data-placement="right" title="Remove target" data-id="[[id]]" data-type="[[type]]" class="btn btn-danger btn-xs glyphicon glyphicon-remove"></button>
        [[name]]
      </span>
    </div>


  </body>
</html>
